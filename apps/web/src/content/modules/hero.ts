import type { HeroContent } from '@/content/types'
import { SITE_LINKS } from '@/content/modules/site'

export const heroContent: HeroContent = {
  badge: '企业 AI 基础设施 · 全新发布 v2.0',
  title: '企业级大模型智能管理平台',
  subtitle: '聚合调度 · 统一掌控',
  quote: '让每一次 AI 调用,都令人愉悦',
  primaryCta: { label: '申请产品演示', href: SITE_LINKS.demo },
  secondaryCta: { label: '观看产品视频', href: SITE_LINKS.video },
  trustLabel: '已被数百家企业 AI 团队信任',
  trustBrands: ['字节跳动', '美团', '小红书', '招商银行', '中国平安', '京东', '网易'],
}
