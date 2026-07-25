import { render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import Home from '@/pages/Home'
import type { HomeContent } from '@/content/types'
import { homeContent } from '@/content'

function createHomeFixture(overrides: Partial<HomeContent> = {}): HomeContent {
  return {
    ...homeContent,
    ...overrides,
    hero: {
      ...homeContent.hero,
      title: '测试主标题',
      badge: '测试徽章',
      ...(overrides.hero ?? {}),
    },
  }
}

describe('Home page', () => {
  it('renders injected hero content and major section anchors', async () => {
    const content = createHomeFixture()
    const { container } = render(<Home content={content} />)

    expect(screen.getByRole('heading', { level: 1, name: '测试主标题' })).toBeInTheDocument()
    expect(screen.getByText('测试徽章')).toBeInTheDocument()

    const heroCta = screen
      .getAllByRole('link', { name: content.hero.primaryCta.label })
      .find((link) => link.getAttribute('href') === content.hero.primaryCta.href)
    expect(heroCta).toBeDefined()

    await waitFor(() => {
      for (const id of [
        'challenges',
        'solutions',
        'capabilities',
        'quota',
        'integration',
        'deployment',
        'testimonials',
        'mission',
        'cta',
      ]) {
        expect(container.querySelector(`#${id}`)).not.toBeNull()
      }
    })
  })

  it('uses default homeContent when no content is injected', () => {
    render(<Home />)
    expect(
      screen.getByRole('heading', { level: 1, name: homeContent.hero.title }),
    ).toBeInTheDocument()
  })
})
