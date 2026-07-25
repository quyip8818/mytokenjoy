import type { IntegrationContent } from '@/content/types'

export const integrationContent: IntegrationContent = {
  title: '全面兼容 OpenAI / Anthropic 接口规范,最小化接入改造成本',
  before: {
    label: '原始直连方式',
  },
  after: {
    label: '接入 Tokenjoy',
    highlighted: true,
  },
  features: [
    {
      icon: 'check',
      title: '100% 协议兼容',
      description:
        '全面兼容 OpenAI、Anthropic 等主流大模型厂商的 API 接口规范,可无缝对接现有应用系统与开发生态。',
      accent: 'brand',
    },
    {
      icon: 'branch',
      title: '业务代码零改造',
      description:
        '无需引入新的 SDK 或重构现有业务逻辑,开发人员仅需修改环境变量配置,即可完成网关切换。',
      accent: 'cyan',
    },
    {
      icon: 'layers',
      title: '完整支持高级特性',
      description:
        '完整支持 Streaming 流式输出、Function Calling、Vision 多模态等高级接口特性,保障调用体验与原厂商完全一致。',
      accent: 'orange',
    },
  ],
}
