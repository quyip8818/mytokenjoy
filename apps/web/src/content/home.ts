import { capabilitiesContent } from '@/content/modules/capabilities'
import { challengesContent } from '@/content/modules/challenges'
import { ctaContent } from '@/content/modules/cta'
import { deploymentContent } from '@/content/modules/deployment'
import { footerContent } from '@/content/modules/footer'
import { heroContent } from '@/content/modules/hero'
import { integrationContent } from '@/content/modules/integration'
import { missionContent } from '@/content/modules/mission'
import { navContent } from '@/content/modules/nav'
import { quotaContent } from '@/content/modules/quota'
import { solutionsContent } from '@/content/modules/solutions'
import { testimonialsContent } from '@/content/modules/testimonials'
import type { HomeContent } from '@/content/types'

export const homeContent: HomeContent = {
  nav: navContent,
  hero: heroContent,
  challenges: challengesContent,
  solutions: solutionsContent,
  capabilities: capabilitiesContent,
  quota: quotaContent,
  integration: integrationContent,
  deployment: deploymentContent,
  testimonials: testimonialsContent,
  mission: missionContent,
  cta: ctaContent,
  footer: footerContent,
}
