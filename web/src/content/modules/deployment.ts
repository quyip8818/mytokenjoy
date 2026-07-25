import type { DeploymentContent } from '@/content/types'

export const deploymentContent: DeploymentContent = {
  title: '灵活的部署模式,满足不同合规需求',
  modes: [
    {
      id: 'saas',
      accent: 'cyan',
      icon: 'cloud',
      title: 'SaaS 托管模式',
      subtitle: '开箱即用,免运维',
      points: [
        {
          label: '极速接入',
          desc: '无需部署服务器,注册即用,最快 5 分钟完成企业级 AI 平台搭建。',
        },
        {
          label: '无缝升级',
          desc: '平台新功能、新模型自动热更新,始终保持最新体验。',
        },
        {
          label: '弹性扩容',
          desc: '依托云原生架构,从容应对突发流量高峰。',
        },
        {
          label: '适用场景',
          desc: '中小企业、创新团队、对数据绝对隔离要求不高的通用业务场景。',
        },
      ],
    },
    {
      id: 'private',
      accent: 'brand',
      icon: 'server',
      title: '私有化部署模式',
      subtitle: '数据绝对掌控,最高安全合规',
      points: [
        {
          label: '云上 VPC 内网部署',
          desc: '支持在公有云的独立 VPC 环境下部署,业务调用不出公网。',
          tag: '网络隔离·数据不出域',
        },
        {
          label: '本地 IDC 机房部署',
          desc: '支持部署于企业自建物理机房或私有云,满足极高安全审计要求。',
          tag: '物理隔离·绝对合规',
        },
        {
          label: '定制化集成',
          desc: '可深度对接企业内部的 SSO、LDAP 及自有审计系统。',
        },
        {
          label: '适用场景',
          desc: '金融机构、医疗健康、政务单位及大型集团。',
        },
      ],
    },
  ],
}
