import type { SolutionsContent } from '@/content/types'

export const solutionsContent: SolutionsContent = {
  badge: 'Tokenjoy 解决方案',
  title: '直面挑战,构建',
  titleHighlight: '企业 AI 管理闭环',
  subtitle: '针对企业 AI 规模化应用中的核心痛点,Tokenjoy 提供完整的管理与治理能力。',
  learnMoreLabel: '了解更多',
  items: [
    {
      icon: 'wallet',
      title: '精细化预算追踪',
      description:
        '多维度预算管理,实时成本归因。事前设定预算阈值,事中触发预警拦截,事后自动生成多维度成本报表。让每一笔 AI 开销都清晰可控。',
      accent: 'brand',
      points: ['按团队/项目分账', '实时预算预警', '多维度成本报表'],
      href: '#',
    },
    {
      icon: 'shield',
      title: '统一合规审计',
      description:
        '全量 API 调用日志记录,支持敏感词检测、数据脱敏、调用溯源。满足金融、医疗、政务等高合规场景要求。',
      accent: 'cyan',
      points: ['完整调用审计', '敏感数据脱敏', '支持等保合规'],
      href: '#',
    },
    {
      icon: 'activity',
      title: '系统化可观测',
      description:
        '实时监控调用量、Token 消耗、模型响应时长、错误率等关键指标。提供多维看板和告警通知,让 AI 系统运行状态尽在掌握。',
      accent: 'orange',
      points: ['实时调用监控', 'Token 消耗分析', '智能告警通知'],
      href: '#',
    },
  ],
}
