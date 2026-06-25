import { ROUTES } from '@/config/routes'

export const DEMO_CTA_IDS = {
  CREDENTIAL: 'demo-cta-credential',
  IMPORT: 'demo-cta-import',
  BUDGET: 'demo-cta-budget',
  MODEL: 'demo-cta-model',
  OVERRUN: 'demo-cta-overrun',
  CREATE_KEY: 'demo-cta-create-key',
  APPLY_QUOTA: 'demo-cta-apply-quota',
} as const

export type DemoCtaKey = keyof typeof DEMO_CTA_IDS
export type DemoCtaId = (typeof DEMO_CTA_IDS)[DemoCtaKey]

export interface DemoGuideStep {
  id: string
  title: string
  description: string
  path: string
  ctaId?: DemoCtaId
  memberId?: string
}

export const DEMO_GUIDE_STORAGE_KEY = 'tokenjoy-demo-guide-completed'

export const DEMO_GUIDE_STEPS: DemoGuideStep[] = [
  {
    id: 'credential',
    title: '配置凭证',
    description: '连接飞书/钉钉/企业微信数据源',
    path: ROUTES.orgDataSource,
    ctaId: DEMO_CTA_IDS.CREDENTIAL,
    memberId: 'm-admin',
  },
  {
    id: 'import',
    title: '全量导入组织',
    description: '导入部门与成员到平台',
    path: ROUTES.orgDataSource,
    ctaId: DEMO_CTA_IDS.IMPORT,
    memberId: 'm-admin',
  },
  {
    id: 'structure',
    title: '查看组织架构',
    description: '确认导入的部门树与成员',
    path: ROUTES.orgStructure,
    memberId: 'm-admin',
  },
  {
    id: 'budget',
    title: '分配部门预算',
    description: '在预算总览为节点设置预算与预留池',
    path: ROUTES.budgetOverview,
    ctaId: DEMO_CTA_IDS.BUDGET,
    memberId: 'm-admin',
  },
  {
    id: 'model',
    title: '添加自定义模型',
    description: '扩展企业可用模型列表',
    path: ROUTES.modelsList,
    ctaId: DEMO_CTA_IDS.MODEL,
    memberId: 'm-admin',
  },
  {
    id: 'whitelist',
    title: '配置模型白名单',
    description: '为部门勾选可用模型',
    path: ROUTES.modelsRouting,
    memberId: 'm-admin',
  },
  {
    id: 'overrun',
    title: '设置超限策略',
    description: '配置全局预警阈值与阻断规则',
    path: ROUTES.budgetAlerts,
    ctaId: DEMO_CTA_IDS.OVERRUN,
    memberId: 'm-admin',
  },
  {
    id: 'create-key',
    title: '创建 Platform Key',
    description: '切换为成员视角，创建个人 Key',
    path: ROUTES.keysMine,
    ctaId: DEMO_CTA_IDS.CREATE_KEY,
    memberId: 'm-1',
  },
  {
    id: 'apply-quota',
    title: '申请额度追加',
    description: '成员发起额度申请',
    path: ROUTES.keysMine,
    ctaId: DEMO_CTA_IDS.APPLY_QUOTA,
    memberId: 'm-1',
  },
  {
    id: 'approve',
    title: '审批申请',
    description: '切换为李四视角，在审批中心处理',
    path: ROUTES.keysApproval,
    memberId: 'm-2',
  },
]
