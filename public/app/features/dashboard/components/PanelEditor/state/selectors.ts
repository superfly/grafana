import memoizeOne from 'memoize-one';

import { PanelPlugin } from '@grafana/data';
import { getConfig } from 'app/core/config';
import { contextSrv } from 'app/core/services/context_srv';
import {getRulesPermissions, provisioningPermissions} from 'app/features/alerting/unified/utils/access-control';
import { GRAFANA_RULES_SOURCE_NAME } from 'app/features/alerting/unified/utils/datasource';

import { PanelEditorTab, PanelEditorTabId } from '../types';

export const getPanelEditorTabs = memoizeOne((tab?: string, plugin?: PanelPlugin) => {
  const tabs: PanelEditorTab[] = [];

  if (!plugin) {
    return tabs;
  }

  let defaultTab = PanelEditorTabId.Visualize;

  if (plugin.meta.skipDataQuery) {
    return [];
  }

  if (!plugin.meta.skipDataQuery) {
    defaultTab = PanelEditorTabId.Query;

    tabs.push({
      id: PanelEditorTabId.Query,
      text: 'Query',
      icon: 'database',
      active: false,
    });

    tabs.push({
      id: PanelEditorTabId.Transform,
      text: 'Transform data',
      icon: 'process',
      active: false,
    });
  }

  if (shouldShowAlertingTab(plugin)) {
    tabs.push({
      id: PanelEditorTabId.Alert,
      text: 'Alert',
      icon: 'bell',
      active: false,
    });
  }

  const activeTab = tabs.find((item) => item.id === (tab || defaultTab)) ?? tabs[0];
  activeTab.active = true;

  return tabs;
});

export function shouldShowAlertingTab(plugin: PanelPlugin) {
  const { alertingEnabled, unifiedAlertingEnabled } = getConfig();
  const hasRuleReadPermissions = contextSrv.hasPermission(getRulesPermissions(GRAFANA_RULES_SOURCE_NAME).read);
  const isAlertingAvailable = alertingEnabled || (unifiedAlertingEnabled && hasRuleReadPermissions);

  const isGraph = plugin.meta.id === 'graph';
  const isTimeseries = plugin.meta.id === 'timeseries';

  const hasProvisioningPermissions = contextSrv.hasPermission(provisioningPermissions.write);

  return ((isAlertingAvailable && isGraph) || isTimeseries) && hasProvisioningPermissions;
}
