import { request, buildQuery } from './client'
import type { LoginActivityResponse, Profile } from './types'
import type { MyDashboardView } from './types/mydashboard'

export const meApi = {
  // --- profile ---
  getProfile: () => request<Profile>('/me/profile'),

  updateProfile: (params: { name?: string; avatar?: string; alias?: string }) =>
    request<void>('/me/profile', {
      method: 'PUT',
      body: JSON.stringify(params),
    }),

  // --- security ---
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

  getLoginActivity: (params: { limit?: number; offset?: number } = {}) =>
    request<LoginActivityResponse>(`/me/login-activity${buildQuery(params)}`),

  // --- analytics ---
  getDashboard: () => request<MyDashboardView>('/me/dashboard'),
}
