import React, { useState, useEffect } from 'react';

import { CoreApp } from '@grafana/data';
import { LoadingPlaceholder, Modal } from '@grafana/ui';
import { LocalStorageValueProvider } from 'app/core/components/LocalStorageValueProvider';

import { LokiLabelBrowser } from '../../components/LokiLabelBrowser';
import { LokiDatasource } from '../../datasource';
import { LokiQuery } from '../../types';

export interface Props {
  isOpen: boolean;
  datasource: LokiDatasource;
  query: LokiQuery;
  app?: CoreApp;
  onClose: () => void;
  onChange: (query: LokiQuery) => void;
  onRunQuery: () => void;
}

export const LabelBrowserModal = (props: Props) => {
  const { isOpen, onClose, datasource, app } = props;
  const [labelsLoaded, setLabelsLoaded] = useState(false);
  const [hasLogLabels, setHasLogLabels] = useState(false);
  const LAST_USED_LABELS_KEY = 'grafana.datasources.loki.browser.labels';

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    datasource.languageProvider.start().then(() => {
      setLabelsLoaded(true);
      setHasLogLabels(datasource.languageProvider.getLabelKeys().length > 0);
    });
  }, [datasource, isOpen]);

  const changeQuery = (value: string) => {
    const { query, onChange, onRunQuery } = props;
    const nextQuery = { ...query, expr: value };
    onChange(nextQuery);
    onRunQuery();
  };

  const onChange = (selector: string) => {
    changeQuery(selector);
    onClose();
  };

  return (
    <Modal isOpen={isOpen} title="Label browser" onDismiss={onClose}>
      {!labelsLoaded && <LoadingPlaceholder text="Loading labels..." />}
      {labelsLoaded && !hasLogLabels && <p>No labels found.</p>}
      {labelsLoaded && hasLogLabels && (
        <LocalStorageValueProvider<string[]> storageKey={LAST_USED_LABELS_KEY} defaultValue={[]}>
          {(lastUsedLabels, onLastUsedLabelsSave, onLastUsedLabelsDelete) => {
            return (
              <LokiLabelBrowser
                languageProvider={datasource.languageProvider}
                onChange={onChange}
                lastUsedLabels={lastUsedLabels}
                storeLastUsedLabels={onLastUsedLabelsSave}
                deleteLastUsedLabels={onLastUsedLabelsDelete}
                app={app}
              />
            );
          }}
        </LocalStorageValueProvider>
      )}
    </Modal>
  );
};
