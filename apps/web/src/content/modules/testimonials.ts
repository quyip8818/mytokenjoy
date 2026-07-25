import type { TestimonialsContent } from '@/content/types'

export const testimonialsContent: TestimonialsContent = {
  title: '客户说',
  titleHighlight: ' Tokenjoy',
  subtitle: '来自不同行业企业的真实反馈,见证 Tokenjoy 在生产环境中的价值。',
  stats: [
    { value: '500+', label: '企业客户' },
    { value: '10亿+', label: '日均 Token 处理' },
    { value: '99.99%', label: '服务可用性' },
    { value: '40%', label: '平均成本降低' },
  ],
  items: [
    {
      name: '王晓明',
      role: '某互联网大厂 AI 平台负责人',
      quote:
        'Tokenjoy 帮我们解决了多团队 LLM 调用的成本归因难题,现在每个项目的 AI 投入都清晰可见,预算管控效率提升 5 倍。',
    },
    {
      name: '李婷',
      role: '某金融机构 AI 治理专家',
      quote:
        '在金融行业合规要求下,Tokenjoy 的审计与脱敏能力为我们提供了完整的合规闭环,顺利通过了等保三级测评。',
    },
    {
      name: '张工',
      role: '某 SaaS 公司 CTO',
      quote:
        '接入 Tokenjoy 后,我们的 AI 服务 P99 延迟降低 60%,异常告警响应时间从小时级缩短到分钟级,运维效率大幅提升。',
    },
  ],
}
