import { request } from './client'
import type { KeyApproval, Paginated, PlatformKey, ProviderKey } from './types'

export const providerKeyApi = {
  list: () => request<ProviderKey[]>('/keys/provider'),
  create: (data: { provider: string; name: string; key: string }) =>
    request<ProviderKey>('/keys/provider', { method: 'POST', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<void>(`/keys/provider/${id}/toggle`, { method: 'PUT', body: JSON.stringify({ enabled }) }),
  rotate: (id: string, newKey: string) =>
    request<ProviderKey>(`/keys/provider/${id}/rotate`, { method: 'POST', body: JSON.stringify({ newKey }) }),
  delete: (id: string) =>
    request<void>(`/keys/provider/${id}`, { method: 'DELETE' }),
}

export const platformKeyApi = {
  list: (params?: { departmentId?: string; type?: 'member' | 'project'; page?: number; pageSize?: number }) =>
    request<Paginated<PlatformKey>>(`/keys/platform?${new URLSearchParams(params as Record<string, string>)}`),
  create: (data: { name: string; type: 'member' | 'project'; memberId?: string; projectId?: string; departmentId: string; quotaMode: string; quota: number; modelWhitelist: string[]; expiresAt?: string }) =>
    request<PlatformKey>('/keys/platform', { method: 'POST', body: JSON.stringify(data) }),
  revoke: (id: string) =>
    request<void>(`/keys/platform/${id}/revoke`, { method: 'PUT' }),
  delete: (id: string) =>
    request<void>(`/keys/platform/${id}`, { method: 'DELETE' }),
}

export const approvalApi = {
  list: (status?: string) =>
    request<KeyApproval[]>(`/keys/approvals${status ? `?status=${status}` : ''}`),
  approve: (id: string) =>
    request<void>(`/keys/approvals/${id}/approve`, { method: 'PUT' }),
  reject: (id: string, reason?: string) =>
    request<void>(`/keys/approvals/${id}/reject`, { method: 'PUT', body: JSON.stringify({ reason }) }),
}
