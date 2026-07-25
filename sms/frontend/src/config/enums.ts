export type BadgeVariant = 'default' | 'success' | 'warning' | 'destructive' | 'outline'

interface EnumItem {
  label: string
  variant: BadgeVariant
}

export const SUPPLIER_STATUS: Record<string, EnumItem> = {
  potential: { label: '潜在', variant: 'outline' },
  active: { label: '合作中', variant: 'success' },
  frozen: { label: '冻结', variant: 'warning' },
  blacklisted: { label: '黑名单', variant: 'destructive' },
}

export const CONTRACT_STATUS: Record<string, EnumItem> = {
  draft: { label: '草稿', variant: 'outline' },
  active: { label: '生效中', variant: 'success' },
  expired: { label: '已到期', variant: 'warning' },
  terminated: { label: '已终止', variant: 'destructive' },
}

export const ORDER_STATUS: Record<string, EnumItem> = {
  pending: { label: '待审批', variant: 'warning' },
  approved: { label: '已审批', variant: 'default' },
  delivered: { label: '已交付', variant: 'default' },
  completed: { label: '已完成', variant: 'success' },
  cancelled: { label: '已取消', variant: 'outline' },
}

export const MODEL_STATUS: Record<string, EnumItem> = {
  available: { label: '可用', variant: 'success' },
  deprecated: { label: '已下线', variant: 'outline' },
}

export const EVAL_GRADE: Record<string, EnumItem> = {
  A: { label: 'A 优秀', variant: 'success' },
  B: { label: 'B 良好', variant: 'default' },
  C: { label: 'C 合格', variant: 'warning' },
  D: { label: 'D 不合格', variant: 'destructive' },
}

export const MODEL_TYPES = ['文本', '图像', '语音', '多模态', '嵌入'] as const
export const CATEGORIES = ['国内厂商', '国外厂商'] as const
export const DIMENSIONS: Record<string, string> = {
  quality: '模型质量',
  performance: '响应性能',
  price: '价格成本',
  service: '服务支持',
  compliance: '合规安全',
}
