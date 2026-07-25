import type { NavContent } from '@/content/types'
import { SITE_LINKS } from '@/content/modules/site'

export const navContent: NavContent = {
  links: [
    { label: '产品', href: '#challenges' },
    { label: '解决方案', href: '#solutions' },
    { label: '核心能力', href: '#capabilities' },
    { label: '预算管控', href: '#quota' },
    { label: '快速接入', href: '#integration' },
    { label: '部署模式', href: '#deployment' },
    { label: '关于我们', href: '#mission' },
  ],
  homeHref: SITE_LINKS.home,
  loginHref: SITE_LINKS.login,
  loginLabel: '登录',
  ctaHref: SITE_LINKS.demo,
  ctaLabel: '申请演示',
}
