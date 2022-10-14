package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/sqlstore/commonSession"
	"github.com/grafana/grafana/pkg/services/sqlstore/sqlxsession"
)

var tsclogger = log.New("sqlstore.transactions")

// DBSessionTx is a type alias to DBSession in order to reduce XORM ambigiuty between sessions and transactions
type DBSessionTx = DBSession

func (tx *DBSessionTx) ConcreteType() *DBSessionTx {
	return tx
}

// WithTransactionalDbSession calls the callback with a session within a transaction.
func (ss *SQLStore) WithTransactionalDbSession(ctx context.Context, callback DBTransactionFunc) error {
	return inTransactionWithRetryCtx[*DBSessionTx](ctx, &XormEngine{ss.engine}, ss.bus, callback, 0)
}

// InTransaction starts a transaction and calls the fn
// It stores the session in the context
func (ss *SQLStore) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return ss.inTransactionWithRetry(ctx, fn, 0)
}

func (ss *SQLStore) SqlxInTransactionWithRetry(ctx context.Context, fn func(ctx context.Context) error, retry int) error {
	return inTransactionWithRetryCtx[*sqlxsession.DBSessionTx](ctx, ss.GetSqlxSession(), ss.bus, func(sess commonSession.Tx[*sqlxsession.DBSessionTx]) error {
		withValue := context.WithValue(ctx, sqlxsession.ContextSQLxTransactionKey{}, sess)
		return fn(withValue)
	}, retry)
}

func (ss *SQLStore) inTransactionWithRetry(ctx context.Context, fn func(ctx context.Context) error, retry int) error {
	return inTransactionWithRetryCtx[*DBSessionTx](ctx, &XormEngine{ss.engine}, ss.bus, func(sess commonSession.Tx[*DBSessionTx]) error {
		withValue := context.WithValue(ctx, ContextSessionKey{}, sess)
		return fn(withValue)
	}, retry)
}

func inTransactionWithRetryCtx[T *DBSessionTx | *sqlxsession.DBSessionTx](ctx context.Context, engine commonSession.TxSessionGetter[T], bus bus.Bus, callback func(commonSession.Tx[T]) error, retry int) error {
	sess, isNew, err := engine.StartSessionOrUseExisting(ctx, true)
	if err != nil {
		return err
	}

	if !sess.IsTransactionOpen() && !isNew {
		// this should not happen because the only place that creates reusable session begins a new transaction.
		return fmt.Errorf("cannot reuse existing session that did not start transaction")
	}

	if isNew { // if this call initiated the session, it should be responsible for closing it.
		defer sess.Close()
	}

	err = callback(sess)

	ctxLogger := tsclogger.FromContext(ctx)

	if !isNew {
		ctxLogger.Debug("skip committing the transaction because it belongs to a session created in the outer scope")
		// Do not commit the transaction if the session was reused.
		return err
	}

	// special handling of database locked errors for sqlite, then we can retry 5 times
	var sqlError sqlite3.Error
	if errors.As(err, &sqlError) && retry < 5 && (sqlError.Code == sqlite3.ErrLocked || sqlError.Code == sqlite3.ErrBusy) {
		if rollErr := sess.Rollback(); rollErr != nil {
			return fmt.Errorf("rolling back transaction due to error failed: %s: %w", rollErr, err)
		}

		time.Sleep(time.Millisecond * time.Duration(10))
		ctxLogger.Info("Database locked, sleeping then retrying", "error", err, "retry", retry)
		return inTransactionWithRetryCtx(ctx, engine, bus, callback, retry+1)
	}

	if err != nil {
		if rollErr := sess.Rollback(); rollErr != nil {
			return fmt.Errorf("rolling back transaction due to error failed: %s: %w", rollErr, err)
		}
		return err
	}
	if err := sess.Commit(); err != nil {
		return err
	}

	events := sess.GetEvents()
	if len(events) > 0 {
		for _, e := range events {
			if err = bus.Publish(ctx, e); err != nil {
				ctxLogger.Error("Failed to publish event after commit.", "error", err)
			}
		}
	}

	return nil
}

func (ss *SQLStore) SQLxInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if ss.Cfg.IsFeatureToggleEnabled("newDBLibrary") {
		return ss.SqlxInTransactionWithRetry(ctx, fn, 0)
	}
	return ss.inTransactionWithRetry(ctx, fn, 0)
}
