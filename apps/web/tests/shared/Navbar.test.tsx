import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Navbar } from '@/shared/Navbar'
import type { NavContent } from '@/content/types'

const navFixture: NavContent = {
  links: [
    { label: '产品', href: '#challenges' },
    { label: '解决方案', href: '#solutions' },
  ],
  homeHref: '#',
  loginHref: '#login',
  loginLabel: '登录',
  ctaHref: '#cta',
  ctaLabel: '申请演示',
}

describe('Navbar', () => {
  it('renders injected nav labels and CTA', () => {
    render(<Navbar content={navFixture} />)

    expect(screen.getByRole('link', { name: /Tokenjoy/i })).toHaveAttribute('href', '#')
    expect(screen.getAllByRole('link', { name: '产品' }).length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: '申请演示' }).length).toBeGreaterThan(0)
  })

  it('opens mobile menu via toggle', () => {
    render(<Navbar content={navFixture} />)

    fireEvent.click(screen.getByRole('button', { name: 'Toggle menu' }))
    expect(screen.getByRole('button', { name: 'Toggle menu' })).toHaveAttribute(
      'aria-expanded',
      'true',
    )
    expect(screen.getAllByRole('link', { name: '解决方案' }).length).toBeGreaterThan(1)
  })
})
