import type { ChallengesContent } from '@/content/types'

export const challengesContent: ChallengesContent = {
  title: '企业 AI 规模化应用面临的三大核心挑战',
  subtitle:
    '在企业 AI 应用从 PoC 走向规模化生产的过程中,以下三个核心问题成为阻碍业务价值释放的关键瓶颈。',
  items: [
    {
      id: '01',
      title: '预算管控缺失',
      accent: 'brand',
      icon: 'budget',
      description:
        '多团队共享 LLM 调用资源,成本归因粒度粗放,无法实现按团队、项目维度的**精细化预算追踪**。费用超支往往在月度结算后方可发现,缺乏事前预警与事中管控机制。',
    },
    {
      id: '02',
      title: '安全合规风险',
      accent: 'cyan',
      icon: 'security',
      description:
        '企业内部 API 调用行为缺乏统一审计,**敏感数据外传风险**难以管控。调用主体、模型来源及输入内容均无法溯源,无法满足监管合规及内部审计的基本要求。',
    },
    {
      id: '03',
      title: '运营可观测性不足',
      accent: 'orange',
      icon: 'observability',
      description:
        '调用量、Token 消耗及模型服务质量缺乏**系统性监控机制**,管理层无法获取有效数据支撑 AI 资源投入决策,资源利用率低下,优化方向不明。',
    },
  ],
}
