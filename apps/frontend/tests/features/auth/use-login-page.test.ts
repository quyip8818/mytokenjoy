import { describe, expect, it, vi } from 'vitest'
import { useLoginPage } from '@/features/auth/hooks/use-login-page'
import { renderHookWithProviders } from '@tests/utils'

const { login, refreshSession } = vi.hoisted(() => ({
  login: vi.fn(),
  refreshSession: vi.fn(),
}))

vi.mock('@/api/auth', () => ({
  authApi: { login, logout: vi.fn() },
}))

vi.mock('@/features/session', () => ({
  useSession: () => ({ refreshSession }),
}))

describe('useLoginPage', () => {
  it('returns login form state and handlers', () => {
    const { result } = renderHookWithProviders(() => useLoginPage(), {
      initialEntries: ['/login'],
    })

    expect(typeof result.current.handleSubmit).toBe('function')
    expect(typeof result.current.setEmail).toBe('function')
    expect(typeof result.current.setPassword).toBe('function')
    expect(typeof result.current.setCompanyId).toBe('function')
    expect(result.current.submitting).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('prefills company id from query string', () => {
    const { result } = renderHookWithProviders(() => useLoginPage(), {
      initialEntries: ['/login?companyid=00000000-0000-7000-8000-000000000002'],
    })

    expect(result.current.companyId).toBe('00000000-0000-7000-8000-000000000002')
  })

  it('sends companyId on login when provided', async () => {
    login.mockResolvedValue({ memberId: 'm-1' })
    refreshSession.mockResolvedValue(undefined)

    const { result } = renderHookWithProviders(() => useLoginPage(), {
      initialEntries: ['/login?companyid=00000000-0000-7000-8000-000000000002'],
    })

    await result.current.handleSubmit({ preventDefault: () => undefined } as React.FormEvent)

    expect(login).toHaveBeenCalledWith(
      expect.objectContaining({
        companyId: '00000000-0000-7000-8000-000000000002',
      }),
    )
  })
})
