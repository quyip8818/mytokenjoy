import { create } from 'zustand'

interface SessionUser {
  id: string
  username: string
  realName: string
  role: string
}

interface SessionState {
  user: SessionUser | null
  setUser: (user: SessionUser | null) => void
  logout: () => void
}

export const useSession = create<SessionState>((set) => ({
  user: (() => {
    const raw = localStorage.getItem('sms_user')
    return raw ? JSON.parse(raw) : null
  })(),
  setUser: (user) => {
    if (user) localStorage.setItem('sms_user', JSON.stringify(user))
    else localStorage.removeItem('sms_user')
    set({ user })
  },
  logout: () => {
    localStorage.removeItem('sms_access_token')
    localStorage.removeItem('sms_user')
    set({ user: null })
    window.location.href = '/login'
  },
}))
