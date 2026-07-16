import { request } from './client'

export interface LoginInput {
  email: string
  password: string
  companyId?: number
}

export const authApi = {
  login: (input: LoginInput) =>
    request<{ memberId: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  logout: () =>
    request<void>('/auth/logout', {
      method: 'POST',
    }),
}
