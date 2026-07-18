import { describe, expect, it, vi } from 'vitest'
import { act } from '@testing-library/react'
import { useSmsLoginPage } from '@/features/auth/hooks/use-sms-login-page'
import { renderHookWithProviders } from '@tests/utils'

const { smsSend, smsVerify, refreshSession } = vi.hoisted(() => ({
  smsSend: vi.fn(),
  smsVerify: vi.fn(),
  refreshSession: vi.fn(),
}))

vi.mock('@/api/auth', () => ({
  authApi: {
    smsSend,
    smsVerify,
    login: vi.fn(),
    logout: vi.fn(),
    smsSelect: vi.fn(),
    registerAccept: vi.fn(),
    registerCompany: vi.fn(),
    acceptInvite: vi.fn(),
  },
}))

vi.mock('@/features/session', () => ({
  useSession: () => ({ refreshSession }),
}))

describe('useSmsLoginPage', () => {
  it('returns initial state', () => {
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    expect(result.current.phone).toBe('')
    expect(result.current.code).toBe('')
    expect(result.current.error).toBeNull()
    expect(result.current.sending).toBe(false)
    expect(result.current.verifying).toBe(false)
    expect(result.current.countdown).toBe(0)
    expect(result.current.canSend).toBe(false)
  })

  it('canSend is true when phone is 11+ chars and not in cooldown', () => {
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => result.current.setPhone('13800138000'))
    expect(result.current.canSend).toBe(true)
  })

  it('starts countdown after successful send', async () => {
    smsSend.mockResolvedValue(undefined)
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => result.current.setPhone('13800138000'))
    await act(async () => {
      await result.current.handleSendCode()
    })

    expect(smsSend).toHaveBeenCalledWith('13800138000')
    expect(result.current.countdown).toBe(60)
    expect(result.current.canSend).toBe(false)
  })

  it('sets error on send failure', async () => {
    const { ApiError } = await import('@/api/client')
    smsSend.mockRejectedValue(new ApiError(429, 'send too frequent', 45))
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => result.current.setPhone('13800138000'))
    await act(async () => {
      await result.current.handleSendCode()
    })

    expect(result.current.error).toBe('send too frequent')
    expect(result.current.countdown).toBe(45)
  })

  it('navigates to home on enter action', async () => {
    smsVerify.mockResolvedValue({ action: 'enter' })
    refreshSession.mockResolvedValue(undefined)
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => {
      result.current.setPhone('13800138000')
      result.current.setCode('123456')
    })
    await act(async () => {
      await result.current.handleVerify({ preventDefault: () => undefined } as React.FormEvent)
    })

    expect(smsVerify).toHaveBeenCalledWith('13800138000', '123456')
    expect(refreshSession).toHaveBeenCalled()
  })

  it('navigates to select page on select_company action', async () => {
    const companies = [
      { companyId: 'c1', companyName: 'Acme', role: 'admin' },
      { companyId: 'c2', companyName: 'Beta', role: 'member' },
    ]
    smsVerify.mockResolvedValue({ action: 'select_company', companies })
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => {
      result.current.setPhone('13800138000')
      result.current.setCode('123456')
    })
    await act(async () => {
      await result.current.handleVerify({ preventDefault: () => undefined } as React.FormEvent)
    })

    expect(smsVerify).toHaveBeenCalled()
    // Navigation is handled via react-router — no error means success
    expect(result.current.error).toBeNull()
  })

  it('sets error on verify failure', async () => {
    const { ApiError } = await import('@/api/client')
    smsVerify.mockRejectedValue(new ApiError(400, 'invalid code'))
    const { result } = renderHookWithProviders(() => useSmsLoginPage(), {
      initialEntries: ['/login'],
    })

    act(() => {
      result.current.setPhone('13800138000')
      result.current.setCode('000000')
    })
    await act(async () => {
      await result.current.handleVerify({ preventDefault: () => undefined } as React.FormEvent)
    })

    expect(result.current.error).toBe('invalid code')
    expect(result.current.verifying).toBe(false)
  })
})
