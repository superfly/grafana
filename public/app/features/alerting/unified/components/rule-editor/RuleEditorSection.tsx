import { css, cx } from '@emotion/css';
import React, { ReactElement } from 'react';

import { GrafanaTheme2 } from '@grafana/data';
import { FieldSet, Text, useStyles2, Stack } from '@grafana/ui';

export interface RuleEditorSectionProps {
  title: string;
  stepNo: number;
  description?: string | ReactElement;
  fullWidth?: boolean;
}

export const RuleEditorSection = ({
  title,
  stepNo,
  children,
  fullWidth = false,
  description,
}: React.PropsWithChildren<RuleEditorSectionProps>) => {
  const styles = useStyles2(getStyles);

  return (
    <div className={styles.parent}>
      <FieldSet
        className={cx(fullWidth && styles.fullWidth)}
      >
        <Stack direction="column">
          {children}
        </Stack>
      </FieldSet>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  parent: css`
    display: flex;
    flex-direction: row;
    max-width: ${theme.breakpoints.values.xl}px;
    border: solid 1px ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(2)} ${theme.spacing(3)};
  `,
  description: css`
    margin-top: -${theme.spacing(2)};
  `,
  fullWidth: css`
    width: 100%;
  `,
});
