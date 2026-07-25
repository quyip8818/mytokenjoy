import { describe, it, expect, beforeEach } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { useSession } from '@/features/session'

describe('useSession', () => {
  beforeEach(() => {
    localStorage.clear()
    // reset zustand store
    useSession.setState({ user: null })
  })

  it('initial state is null when no stored user', () => {
    const { result } = renderHook(() => useSession())
    expect(result.current.user).toBeNull()
  })

  it('setUser persists to localStorage', () => {
    const { result } = renderHook(() => useSession())
    const user = { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' }

    act(() => result.current.setUser(user))

    expect(result.current.user).toEqual(user)
    expect(JSON.parse(localStorage.getItem('sms_user')!)).toEqual(user)
  })

  it('setUser(null) removes from localStorage', () => {
    const { result } = renderHook(() => useSession())
    const user = { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' }

    act(() => result.current.setUser(user))
    act(() => result.current.setUser(null))

    expect(result.current.user).toBeNull()
    expect(localStorage.getItem('sms_user')).toBeNull()
  })

  it('logout clears tokens and user', () => {
    const { result } = renderHook(() => useSession())
    localStorage.setItem('sms_access_token', 'tok')
    act(() =>
      result.current.setUser({ id: 'u1', username: 'admin', realName: '管理员', role: 'admin' }),
    )

    // jsdom doesn't support navigation; just verify state/storage cleanup
    delete (window as any).location
    ;(window as any).location = { href: '' }

    act(() => result.current.logout())

    expect(result.current.user).toBeNull()
    expect(localStorage.getItem('sms_access_token')).toBeNull()
    expect(localStorage.getItem('sms_user')).toBeNull()
  })
})
