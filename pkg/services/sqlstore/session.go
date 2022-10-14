package sqlstore

import (
	"context"
	"reflect"

	"xorm.io/xorm"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/sqlstore/commonSession"
)

var sessionLogger = log.New("sqlstore.session")

type DBSession struct {
	*xorm.Session
	transactionOpen bool
	events          []interface{}
}

type DBTransactionFunc func(sess commonSession.Tx[*DBSessionTx]) error
type DBSessionFunc func(sess *DBSession) error

func (sess *DBSession) publishAfterCommit(msg interface{}) {
	sess.events = append(sess.events, msg)
}

func (sess *DBSession) PublishAfterCommit(msg interface{}) {
	sess.events = append(sess.events, msg)
}

func (sess *DBSession) GetEvents() []interface{} {
	return sess.events
}

type XormEngine struct {
	*xorm.Engine
}

// implements SessionGetter[*DBSessionTx]
func (xe *XormEngine) StartSessionOrUseExisting(ctx context.Context, beginTran bool) (commonSession.Tx[*DBSessionTx], bool, error) {
	return xe.startSessionOrUseExisting(ctx, beginTran)
}

func (xe *XormEngine) startSessionOrUseExisting(ctx context.Context, beginTran bool) (*DBSession, bool, error) {
	value := ctx.Value(ContextSessionKey{})
	var sess *DBSession
	sess, ok := value.(*DBSession)

	if ok {
		ctxLogger := sessionLogger.FromContext(ctx)
		ctxLogger.Debug("reusing existing session", "transaction", sess.transactionOpen)
		sess.Session = sess.Session.Context(ctx)
		return sess, false, nil
	}

	newSess := &DBSession{Session: xe.NewSession(), transactionOpen: beginTran}
	if beginTran {
		err := newSess.Begin()
		if err != nil {
			return nil, false, err
		}
	}

	newSess.Session = newSess.Session.Context(ctx)
	return newSess, true, nil
}

func (sess *DBSession) IsTransactionOpen() bool {
	return sess.transactionOpen
}

// WithDbSession calls the callback with the session in the context (if exists).
// Otherwise it creates a new one that is closed upon completion.
// A session is stored in the context if sqlstore.InTransaction() has been been previously called with the same context (and it's not committed/rolledback yet).
func (ss *SQLStore) WithDbSession(ctx context.Context, callback DBSessionFunc) error {
	return withDbSession(ctx, &XormEngine{ss.engine}, callback)
}

// WithNewDbSession calls the callback with a new session that is closed upon completion.
func (ss *SQLStore) WithNewDbSession(ctx context.Context, callback DBSessionFunc) error {
	sess := &DBSession{Session: ss.engine.NewSession(), transactionOpen: false}
	defer sess.Close()
	return callback(sess)
}

func withDbSession(ctx context.Context, engine *XormEngine, callback DBSessionFunc) error {
	sess, isNew, err := engine.startSessionOrUseExisting(ctx, false)
	if err != nil {
		return err
	}
	if isNew {
		defer sess.Close()
	}
	return callback(sess)
}

func (sess *DBSession) InsertId(bean interface{}) (int64, error) {
	table := sess.DB().Mapper.Obj2Table(getTypeName(bean))

	if err := dialect.PreInsertId(table, sess.Session); err != nil {
		return 0, err
	}
	id, err := sess.Session.InsertOne(bean)
	if err != nil {
		return 0, err
	}
	if err := dialect.PostInsertId(table, sess.Session); err != nil {
		return 0, err
	}

	return id, nil
}

func getTypeName(bean interface{}) (res string) {
	t := reflect.TypeOf(bean)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
