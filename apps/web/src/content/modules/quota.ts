import type { QuotaContent } from '@/content/types'

export const quotaContent: QuotaContent = {
  title: '多层级预算配额与成本硬性约束机制',
  flowTitle: '预算逐级下发机制',
  featuresTitle: '核心管控特性',
  levels: [
    {
      title: '企业总额度池',
      subtitle: '全局预算上限,统一计费结算',
      background: 'linear-gradient(135deg, #1E293B 0%, #0F172A 100%)',
      subtitleClass: 'text-white/60',
    },
    {
      title: '部门/团队额度',
      subtitle: '按业务线独立分配,分别计费',
      background: 'linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%)',
      subtitleClass: 'text-white/80',
    },
    {
      title: '项目额度',
      subtitle: '资源专项专用,防止跨项目资源挤占',
      background: 'linear-gradient(135deg, #22D3EE 0%, #0891B2 100%)',
      subtitleClass: 'text-white/90',
    },
    {
      title: '开发者 Platform Key',
      subtitle: '精确至个人级别的调用凭证',
      background: 'linear-gradient(135deg, #FB923C 0%, #F97316 100%)',
      subtitleClass: 'text-white/90',
    },
  ],
  features: [
    {
      title: '成本硬性约束(Hard Constraint)',
      description:
        '平台底层实现**额度硬性约束机制**。当子层级配额耗尽时,API 调用自动触发熔断,确保各层级资源消耗不超出企业设定的预算上限。',
      accent: 'brand',
    },
    {
      title: '灵活的分配与预警策略',
      description:
        '支持按自然周期(月度/季度)自动重置配额。可配置多级用量预警阈值,通过邮件或 IM 渠道向责任人发送预警通知。',
      accent: 'cyan',
    },
    {
      title: '轻量化审批流',
      description:
        '团队负责人在部门分配的额度范围内,拥有**自主审批授权**。开发人员申请临时提额可由负责人小程序审批即时生效,无需层层流转至 IT 部门,有效保障开发进度。',
      accent: 'orange',
    },
  ],
}
