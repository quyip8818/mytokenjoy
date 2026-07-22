import { request } from './client'

export interface ProfileCompany {
  companyId: string
  companyName: string
  role: string
  current: boolean
}

export interface Profile {
  phone: string
  email: string
  name: string
  hasPassword: boolean
  companies: ProfileCompany[]
}

export const accountApi = {
  getProfile: () => request<Profile>('/me/profile'),

  changePassword: (params: { oldPassword?: string; newPassword: string }) =>
    request<void>('/me/change-password', {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  changePhone: (phone: string, code: string) =>
    request<void>('/me/change-phone', {
      method: 'POST',
      body: JSON.stringify({ phone, code }),
    }),

  changeEmail: (email: string, code: string) =>
    request<void>('/me/change-email', {
      method: 'POST',
      body: JSON.stringify({ email, code }),
    }),

  revokeSessions: () =>
    request<void>('/me/revoke-sessions', {
      method: 'POST',
    }),
}
