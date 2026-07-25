import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { RichText } from '@/shared/RichText'

describe('RichText', () => {
  it('renders plain text without bold markers', () => {
    render(<RichText text="plain description" />)
    expect(screen.getByText('plain description')).toBeInTheDocument()
    expect(screen.queryByRole('strong')).not.toBeInTheDocument()
  })

  it('renders markdown bold segments as strong nodes', () => {
    render(<RichText text="before **重要能力** after" />)
    expect(screen.getByText('重要能力').tagName).toBe('STRONG')
    expect(screen.getByText(/before/)).toBeInTheDocument()
    expect(screen.getByText(/after/)).toBeInTheDocument()
  })
})
