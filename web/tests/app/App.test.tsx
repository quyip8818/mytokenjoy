import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from '@/App'
import { homeContent } from '@/content'

describe('App', () => {
  it('renders the landing page', () => {
    render(<App />)

    expect(
      screen.getByRole('heading', { level: 1, name: homeContent.hero.title }),
    ).toBeInTheDocument()
  })
})
