export const COST_PERIOD = {
  CURRENT_MONTH: 'current_month',
  CURRENT_WEEK: 'current_week',
  LAST_MONTH: 'last_month',
  LAST_7_DAYS: 'last_7_days',
  LAST_30_DAYS: 'last_30_days',
  CUSTOM: 'custom',
} as const

export const COST_PERIOD_LABELS: Record<string, string> = {
  [COST_PERIOD.CURRENT_MONTH]: '本月',
  [COST_PERIOD.CURRENT_WEEK]: '本周',
  [COST_PERIOD.LAST_MONTH]: '上月',
  [COST_PERIOD.LAST_7_DAYS]: '近 7 天',
  [COST_PERIOD.LAST_30_DAYS]: '近 30 天',
  [COST_PERIOD.CUSTOM]: '自定义',
}

/** Preset periods shown in the date range picker (excludes custom). */
export const DATE_RANGE_PRESETS = [
  COST_PERIOD.CURRENT_MONTH,
  COST_PERIOD.CURRENT_WEEK,
  COST_PERIOD.LAST_MONTH,
  COST_PERIOD.LAST_7_DAYS,
  COST_PERIOD.LAST_30_DAYS,
] as const

export const COST_GRANULARITY = {
  DAY: 'day',
  WEEK: 'week',
  MONTH: 'month',
} as const

export const COST_GRANULARITY_LABELS: Record<string, string> = {
  [COST_GRANULARITY.DAY]: '按天',
  [COST_GRANULARITY.WEEK]: '按周',
  [COST_GRANULARITY.MONTH]: '按月',
}

export const MODEL_NOT_IN_DEPT_MESSAGE = '该模型不在您部门的可用范围内'
