import { ArrowRight, Sparkles, Wallet, Zap } from 'lucide-react'
import { hasPermission } from '@/lib/permissions'
import type { AnnouncementConfig } from './types'

/**
 * 公告注册表。
 * ponytail: 新增公告只需往数组里加一条配置。
 * 升级路径：改为远程配置 + A/B testing。
 */
export const announcements: AnnouncementConfig[] = [
  {
    id: 'trial-welcome',
    gradient: 'from-indigo-500 via-purple-500 to-pink-500',
    icon: Sparkles,
    title: '欢迎体验 TokenJoy',
    subtitle: '您的试用环境已准备就绪',
    features: [
      {
        icon: Wallet,
        title: '1000 元模拟资金已到账',
        description: '可用于配置预算、分配额度、测试完整的用量管控流程',
        cardClass: 'border-indigo-100 bg-indigo-50/50',
        iconClass: 'bg-indigo-100 text-indigo-600',
      },
      {
        icon: Zap,
        title: '模拟流量即时验证',
        description: '签发 Key 后可调用 Mock 模型，验证预算消耗和告警功能',
        cardClass: 'border-purple-100 bg-purple-50/50',
        iconClass: 'bg-purple-100 text-purple-600',
      },
      {
        icon: ArrowRight,
        title: '充值即可升级',
        description: '完成充值后自动升级为正式账户，接入真实模型服务',
        cardClass: 'border-emerald-100 bg-emerald-50/50',
        iconClass: 'bg-emerald-100 text-emerald-600',
      },
    ],
    ctaLabel: '开始体验',
    shouldShow: (session) =>
      session.companyType === 'trial' && hasPermission(session.permissions, 'org:admin'),
  },
]
