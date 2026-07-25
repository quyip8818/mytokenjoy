import type { AccentKey } from '@/shared/accents'

export interface NavLink {
  label: string
  href: string
}

export interface NavContent {
  links: NavLink[]
  homeHref: string
  loginHref: string
  loginLabel: string
  ctaHref: string
  ctaLabel: string
}

export interface HeroContent {
  badge: string
  title: string
  subtitle: string
  quote: string
  primaryCta: { label: string; href: string }
  secondaryCta: { label: string; href: string }
  trustLabel: string
  trustBrands: string[]
}

export interface ChallengeItem {
  id: string
  title: string
  accent: AccentKey
  icon: 'budget' | 'security' | 'observability'
  description: string
}

export interface ChallengesContent {
  title: string
  subtitle: string
  items: ChallengeItem[]
}

export interface SolutionItem {
  icon: 'wallet' | 'shield' | 'activity'
  title: string
  description: string
  accent: AccentKey
  points: string[]
  href: string
}

export interface SolutionsContent {
  badge: string
  title: string
  titleHighlight: string
  subtitle: string
  items: SolutionItem[]
  learnMoreLabel: string
}

export interface CapabilityItem {
  icon: 'plug' | 'shield' | 'wallet' | 'activity'
  title: string
  accent: AccentKey
  description: string
  points: string[]
}

export interface CapabilitiesContent {
  title: string
  items: CapabilityItem[]
}

export interface QuotaLevel {
  title: string
  subtitle: string
  background: string
  subtitleClass: string
}

export interface QuotaFeature {
  title: string
  description: string
  accent: AccentKey
}

export interface QuotaContent {
  title: string
  flowTitle: string
  featuresTitle: string
  levels: QuotaLevel[]
  features: QuotaFeature[]
}

export interface CodeSnippet {
  label: string
  highlighted?: boolean
}

export interface IntegrationFeature {
  icon: 'check' | 'branch' | 'layers'
  title: string
  description: string
  accent: AccentKey
}

export interface IntegrationContent {
  title: string
  before: CodeSnippet
  after: CodeSnippet
  features: IntegrationFeature[]
}

export interface DeploymentPoint {
  label: string
  desc: string
  tag?: string
}

export interface DeploymentMode {
  id: string
  accent: AccentKey
  icon: 'cloud' | 'server'
  title: string
  subtitle: string
  points: DeploymentPoint[]
}

export interface DeploymentContent {
  title: string
  modes: DeploymentMode[]
}

export interface StatItem {
  value: string
  label: string
}

export interface TestimonialItem {
  name: string
  role: string
  quote: string
}

export interface TestimonialsContent {
  title: string
  titleHighlight: string
  subtitle: string
  stats: StatItem[]
  items: TestimonialItem[]
}

export interface MissionCard {
  icon: 'eye' | 'users'
  title: string
  content: string
}

export interface MissionContent {
  title: string
  quote: string
  description: string
  cards: MissionCard[]
}

export interface CtaContent {
  title: string
  titleHighlight: string
  description: string
  benefits: string[]
  primaryCta: { label: string; href: string }
  secondaryCta: { label: string; href: string }
}

export interface FooterColumn {
  title: string
  links: { label: string; href: string }[]
}

export interface SocialLink {
  icon: 'github' | 'twitter' | 'linkedin' | 'wechat'
  label: string
  href: string
}

export interface FooterContent {
  description: string
  email: string
  copyright: string
  icpNumber: string
  icpUrl: string
  columns: FooterColumn[]
  socialLinks: SocialLink[]
}

export interface HomeContent {
  nav: NavContent
  hero: HeroContent
  challenges: ChallengesContent
  solutions: SolutionsContent
  capabilities: CapabilitiesContent
  quota: QuotaContent
  integration: IntegrationContent
  deployment: DeploymentContent
  testimonials: TestimonialsContent
  mission: MissionContent
  cta: CtaContent
  footer: FooterContent
}
