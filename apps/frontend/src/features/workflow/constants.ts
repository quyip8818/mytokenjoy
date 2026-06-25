export const QUOTA_INSUFFICIENT_MESSAGE = '额度不足，请先申请追加'

export const WORKFLOW_LAYER_WIDTH = {
  layer1: '75vw',
  layer2: '62vw',
  layer3: '50vw',
} as const

export const WORKFLOW_LAYER_MAX_WIDTH = {
  layer1: 1200,
  layer2: 960,
  layer3: 720,
} as const

export const WORKFLOW_ANIMATION_MS = 280
export const WORKFLOW_MAX_DEPTH = 3
export const WORKFLOW_PEEK_WIDTH_PX = 12

export const WORKFLOW_INFO_BOX_CLASS = 'rounded-lg border border-border/60 bg-muted/50 p-5 text-sm'

export const WORKFLOW_INFO_BOX_CODE_CLASS =
  'rounded-lg border border-border bg-muted/50 px-4 py-3 font-mono text-sm'

export const WORKFLOW_DROPZONE_CLASS =
  'flex w-full flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-border/60 bg-muted/30 py-10 transition-colors hover:border-primary/40 hover:bg-primary/5'

export const WORKFLOW_TABLE_HEAD_CLASS = 'bg-muted/50'

export const WORKFLOW_LIST_ITEM_CLASS =
  'rounded-lg border border-border/50 hover:bg-primary/5 transition-colors'

export const WORKFLOW_LIST_ITEM_SELECTED_CLASS = 'border-primary/30 bg-primary/5'

export const WORKFLOW_SCROLL_LIST_CLASS =
  'max-h-[50vh] overflow-y-auto rounded-lg border border-border/60 divide-y divide-border/40'

export const WORKFLOW_FORM_FIELD_CLASS = 'space-y-1.5'
export const WORKFLOW_FIELD_ERROR_CLASS = 'text-xs text-destructive'
