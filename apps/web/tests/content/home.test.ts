import { describe, expect, it } from 'vitest'
import { homeContent, SITE_LINKS } from '@/content'
import { navContent } from '@/content/modules/nav'

const HOME_SECTION_IDS = [
  'challenges',
  'solutions',
  'capabilities',
  'quota',
  'integration',
  'deployment',
  'testimonials',
  'mission',
  'cta',
] as const

describe('home content', () => {
  it('assembles every homepage content module', () => {
    expect(homeContent.nav).toBe(navContent)
    expect(homeContent.hero.title).toBeTruthy()
    expect(homeContent.challenges.items).toHaveLength(3)
    expect(homeContent.solutions.items).toHaveLength(3)
    expect(homeContent.capabilities.items).toHaveLength(4)
    expect(homeContent.quota.levels).toHaveLength(4)
    expect(homeContent.integration.features).toHaveLength(3)
    expect(homeContent.deployment.modes).toHaveLength(2)
    expect(homeContent.testimonials.items).toHaveLength(3)
    expect(homeContent.mission.cards).toHaveLength(2)
    expect(homeContent.cta.benefits.length).toBeGreaterThan(0)
    expect(homeContent.footer.columns.length).toBeGreaterThan(0)
  })

  it('keeps nav anchors within homepage section ids', () => {
    for (const link of homeContent.nav.links) {
      expect(link.href.startsWith('#')).toBe(true)
      expect(HOME_SECTION_IDS).toContain(link.href.slice(1))
    }
  })

  it('wires site links into nav and hero CTAs', () => {
    expect(homeContent.nav.homeHref).toBe(SITE_LINKS.home)
    expect(homeContent.nav.loginHref).toBe(SITE_LINKS.login)
    expect(homeContent.nav.ctaHref).toBe(SITE_LINKS.demo)
    expect(homeContent.hero.primaryCta.href).toBe(SITE_LINKS.demo)
    expect(homeContent.hero.secondaryCta.href).toBe(SITE_LINKS.video)
    expect(SITE_LINKS.demo).toBe('#cta')
  })
})
