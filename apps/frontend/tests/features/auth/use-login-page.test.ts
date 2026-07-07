import { describe, expect, it } from 'vitest'
import { useLoginPage } from '@/features/auth/hooks/use-login-page'
import { renderHookWithProviders } from '@tests/utils'

describe('useLoginPage', () => {
  it('returns login form state and handlers', () => {
    const { result } = renderHookWithProviders(() => useLoginPage(), {
      initialEntries: ['/login'],
    })

    expect(typeof result.current.handleSubmit).toBe('function')
    expect(typeof result.current.setEmail).toBe('function')
    expect(typeof result.current.setPassword).toBe('function')
    expect(result.current.submitting).toBe(false)
    expect(result.current.error).toBeNull()
  })
})
