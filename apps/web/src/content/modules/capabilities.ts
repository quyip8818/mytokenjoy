import type { CapabilitiesContent } from '@/content/types'

export const capabilitiesContent: CapabilitiesContent = {
  title: '核心能力全景图',
  items: [
    {
      icon: 'plug',
      title: '一站式全模型接入与支付',
      accent: 'cyan',
      description:
        '一个账号,统一接入国内主流大模型(通义千问、文心一言、智谱 GLM、Kimi、豆包等),显著降低多供应商对接的集成成本,实现统一支付与账单管理。',
      points: [
        '一次接入,全企业全模型统一调用',
        '对接单一标准 API,即可调用平台集成的所有模型',
        '统一支付账户,消除多平台分散充值的的管理负担',
      ],
    },
    {
      icon: 'shield',
      title: '高可用与自动容灾',
      accent: 'brand',
      description: '内置智能路由与负载均衡机制,为企业级生产环境提供高可用的服务稳定性保障。',
      points: [
        '模型服务异常时毫秒级自动切换至备用节点',
        '动态负载均衡,支持高并发场景下的弹性调度',
        '保障业务连续性,满足企业级 SLA 可用性要求',
      ],
    },
    {
      icon: 'wallet',
      title: '多层级细粒度预算管控',
      accent: 'orange',
      description: '面向企业多层级组织架构,提供精细化的 Token 资源配置管理与成本归因能力。',
      points: [
        '支持企业 → 部门 → 项目 → 开发者四级配额体系',
        '可配置用量预警阈值与自动熔断策略',
        '资源专项分配,防止跨业务线的资源挤占',
      ],
    },
    {
      icon: 'activity',
      title: '全链路可观测性与审计',
      accent: 'ink',
      description: '提供全链路调用日志与多维度数据看板,满足企业内控审计与安全合规要求。',
      points: [
        '实时监控首 Token 延迟、QPS、TPS 等关键指标',
        '精确至单次请求的全链路调用日志审计',
        '多维度成本分析报表,支持按需导出',
      ],
    },
  ],
}
