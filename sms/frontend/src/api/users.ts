import { request } from './client'
import type { User } from './auth'

export type Role = { code: string; name: string }

const ROLES: Role[] = [
  { code: 'admin', name: '管理员' },
  { code: 'buyer', name: '采购员' },
  { code: 'viewer', name: '观察者' },
]

export const usersApi = {
  list: () => request<User[]>('/users'),
  roles: () => Promise.resolve(ROLES),
  create: (data: {
    username: string
    password: string
    realName: string
    email?: string
    role: string
    status?: number
  }) => request<User>('/users', { method: 'POST', body: JSON.stringify(data) }),
  update: (
    id: string,
    data: { realName: string; email?: string; role: string; password?: string; status?: number },
  ) => request<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/users/${id}`, { method: 'DELETE' }),
}
