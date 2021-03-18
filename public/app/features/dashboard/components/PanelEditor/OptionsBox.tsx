import React, { FC, ReactNode, useState } from 'react';
import { css, cx } from 'emotion';
import _ from 'lodash';
import { GrafanaTheme } from '@grafana/data';
import { Icon, useStyles } from '@grafana/ui';

export interface Props {
  title?: React.ReactNode;
  className?: string;
  children: ReactNode;
}

export const OptionsBox: FC<Props> = ({ title, children, className }) => {
  const [isExpanded, toggleExpand] = useState(true);
  const styles = useStyles(getStyles);

  return (
    <div className={cx(styles.box, className)}>
      {title && (
        <div className={styles.header} onClick={() => toggleExpand(!isExpanded)}>
          <div className={cx(styles.toggle)}>
            <Icon name={isExpanded ? 'angle-down' : 'angle-right'} />
          </div>
          <div className={styles.title}>{title}</div>
        </div>
      )}
      {isExpanded && <div className={styles.body}>{children}</div>}
    </div>
  );
};

const getStyles = (theme: GrafanaTheme) => {
  return {
    box: css`
      padding: 0;
      display: flex;
      flex-direction: column;
      min-height: 0;
      margin-bottom: ${theme.spacing.sm};
    `,
    toggle: css`
      color: ${theme.colors.textWeak};
      margin-right: ${theme.spacing.sm};
    `,
    title: css`
      flex-grow: 1;
      overflow: hidden;
    `,
    header: css`
      display: flex;
      cursor: pointer;
      align-items: baseline;
      padding: ${theme.spacing.xs} ${theme.spacing.sm};
      color: ${theme.colors.formLabel};
      font-weight: ${theme.typography.weight.semibold};
      //background: ${theme.colors.bg2};

      &:hover {
        color: ${theme.colors.text};

        .editor-options-group-toggle {
          color: ${theme.colors.text};
        }
      }
    `,
    body: css`
      position: relative;
      background: ${theme.colors.bodyBg};
      padding: ${theme.spacing.sm} ${theme.spacing.md} 0 ${theme.spacing.md};
    `,
  };
};
