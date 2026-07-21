import { request } from './client'

export interface LoginInput {
  email: string
  password: string
  companyId?: string
}

// --- SMS Auth Types (design doc §5) ---

export interface CompanyOption {
  companyId: string
  companyName: string
  role: string
}

export interface PendingInvite {
  inviteCode: string
  companyId: string
  companyName: string
  role: string
  expiresAt: string
}

export type SmsVerifyResult =
  | { action: 'enter' }
  | { action: 'select_company'; companies: CompanyOption[] }
  | { action: 'choose'; invites: PendingInvite[] }
  | { action: 'not_found' }

export type LoginResult =
  | { memberId: string }
  | { action: 'select_company'; companies: CompanyOption[] }

export const authApi = {
  login: (input: LoginInput) =>
    request<LoginResult>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  logout: () =>
    request<void>('/auth/logout', {
      method: 'POST',
    }),

  // --- SMS endpoints ---
  smsSend: (phone: string) =>
    request<void>('/auth/sms/send', {
      method: 'POST',
      body: JSON.stringify({ phone }),
    }),

  smsVerify: (phone: string, code: string) =>
    request<SmsVerifyResult>('/auth/sms/verify', {
      method: 'POST',
      body: JSON.stringify({ phone, code }),
    }),

  smsSelect: (companyId: string) =>
    request<{ memberId: string; companyId: string }>('/auth/sms/select', {
      method: 'POST',
      body: JSON.stringify({ companyId }),
    }),

  // --- Register endpoints ---
  registerInit: (phone: string, code: string) =>
    request<{ action: 'choose'; invites: PendingInvite[] } | { action: 'login' }>(
      '/auth/register/init',
      {
        method: 'POST',
        body: JSON.stringify({ phone, code }),
      },
    ),

  registerAccept: (inviteCode: string, name: string) =>
    request<{ memberId: string; companyId: string }>('/auth/register/accept', {
      method: 'POST',
      body: JSON.stringify({ inviteCode, name }),
    }),

  registerCompany: (companyName: string, password: string, industry: string, size: string) =>
    request<{ memberId: string; companyId: string }>('/auth/register/company', {
      method: 'POST',
      body: JSON.stringify({ companyName, password, industry, size }),
    }),

  // --- Accept invite (unauthenticated, email link) ---
  acceptInvite: (inviteCode: string, name: string, password: string) =>
    request<{ memberId: string; companyId: string }>('/auth/accept-invite', {
      method: 'POST',
      body: JSON.stringify({ inviteCode, name, password }),
    }),

  // --- Set password (authenticated) ---
  setPassword: (password: string) =>
    request<void>('/auth/set-password', {
      method: 'POST',
      body: JSON.stringify({ password }),
    }),

  // --- Reset password (unauthenticated, SMS verified) ---
  resetPassword: (phone: string, code: string, newPassword: string) =>
    request<void>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ phone, code, newPassword }),
    }),
}
