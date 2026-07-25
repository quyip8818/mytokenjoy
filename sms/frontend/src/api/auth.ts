import { request } from './client'

export interface User {
  id: string
  username: string
  realName: string
  email?: string
  role: string
  status: number
  createdAt: string
  updatedAt: string
}

export const authApi = {
  login: (data: { username: string; password: string }) =>
    request<{ accessToken: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  refresh: () => request<{ accessToken: string }>('/auth/refresh', { method: 'POST' }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  profile: () => request<User>('/auth/profile'),
}
