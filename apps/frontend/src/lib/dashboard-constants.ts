export const COST_PERIOD = {
  CURRENT_MONTH: 'current_month',
  LAST_MONTH: 'last_month',
  LAST_7_DAYS: 'last_7_days',
} as const

export const COST_PERIOD_LABELS: Record<string, string> = {
  [COST_PERIOD.CURRENT_MONTH]: '本月',
  [COST_PERIOD.LAST_MONTH]: '上月',
  [COST_PERIOD.LAST_7_DAYS]: '近 7 天',
}

export const AUDIT_CONTENT_RETENTION_KEY = 'tokenjoy-audit-content-retention'

export const MODEL_NOT_IN_DEPT_MESSAGE = '该模型不在您部门的可用范围内'
