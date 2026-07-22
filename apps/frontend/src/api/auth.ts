import { request } from './client'

export interface LoginInput {
  email: string
  password: string
}

// --- Verify Code Auth Types ---

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

export type VerifyResult =
  | { action: 'enter' }
  | { action: 'select_company'; companies: CompanyOption[] }
  | { action: 'choose'; invites: PendingInvite[] }
  | { action: 'create_company' }
  | { action: 'not_found' }

export type LoginResult =
  | { memberId: string }
  | { action: 'select_company'; companies: CompanyOption[] }
  | { action: 'create_company' }

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

  // --- Verify code endpoints (phone or email) ---
  sendCode: (params: { phone?: string; email?: string; purpose?: string }) =>
    request<void>('/auth/verify-code/send', {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  verifyCode: (params: { phone?: string; email?: string; code: string }) =>
    request<VerifyResult>('/auth/verify-code/verify', {
      method: 'POST',
      body: JSON.stringify(params),
    }),

  selectCompany: (companyId: string) =>
    request<{ memberId: string; companyId: string }>('/auth/select-company', {
      method: 'POST',
      body: JSON.stringify({ companyId }),
    }),

  // --- Register endpoints ---
  registerInit: (identifier: { phone?: string; email?: string }, code: string, password: string, name?: string) =>
    request<{ action: 'choose'; invites: PendingInvite[] } | { action: 'login' }>(
      '/auth/register/init',
      {
        method: 'POST',
        body: JSON.stringify({ ...identifier, code, password, ...(name ? { name } : {}) }),
      },
    ),

  registerAccept: (inviteCode: string, name?: string) =>
    request<{ memberId: string; companyId: string }>('/auth/register/accept', {
      method: 'POST',
      body: JSON.stringify({ inviteCode, ...(name ? { name } : {}) }),
    }),

  registerCompany: (companyName: string, industry: string, size: string, alias?: string, avatar?: string) =>
    request<{ memberId: string; companyId: string }>('/auth/register/company', {
      method: 'POST',
      body: JSON.stringify({ companyName, industry, size, ...(alias ? { alias } : {}), ...(avatar ? { avatar } : {}) }),
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

  // --- Reset password (unauthenticated, verified) ---
  resetPassword: (identifier: { phone?: string; email?: string }, code: string, newPassword: string) =>
    request<void>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ ...identifier, code, newPassword }),
    }),
}
