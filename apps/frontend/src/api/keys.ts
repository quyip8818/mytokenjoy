import { request, buildQuery } from './client'
import type { Paginated, PlatformKey, PlatformKeyScope, ProviderKey } from './types'

export const providerKeyApi = {
  list: () => request<ProviderKey[]>('/keys/provider'),
  create: (data: { provider: string; name: string; key: string }) =>
    request<ProviderKey>('/keys/provider', { method: 'POST', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<void>(`/keys/provider/${id}/toggle`, {
      method: 'PUT',
      body: JSON.stringify({ enabled }),
    }),
  rotate: (id: string, newKey: string) =>
    request<ProviderKey>(`/keys/provider/${id}/rotate`, {
      method: 'POST',
      body: JSON.stringify({ newKey }),
    }),
  delete: (id: string) => request<void>(`/keys/provider/${id}`, { method: 'DELETE' }),
}

export const platformKeyApi = {
  list: (params?: {
    page?: number
    pageSize?: number
    memberId?: string
    projectId?: string
    departmentId?: string
    scope?: PlatformKeyScope
  }) => request<Paginated<PlatformKey>>(`/keys/platform${buildQuery(params ?? {})}`),
  create: (data: {
    name: string
    scope: PlatformKeyScope
    memberId?: string
    projectId?: string
    budget: number
    modelWhitelist: string[]
  }) => request<PlatformKey>('/keys/platform', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: { name?: string; budget?: number; modelWhitelist?: string[] }) =>
    request<PlatformKey>(`/keys/platform/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  toggle: (id: string, enabled: boolean) =>
    request<PlatformKey>(`/keys/platform/${id}/toggle`, {
      method: 'PUT',
      body: JSON.stringify({ enabled }),
    }),
  rotate: (id: string) => request<PlatformKey>(`/keys/platform/${id}/rotate`, { method: 'POST' }),
  revoke: (id: string) => request<void>(`/keys/platform/${id}/revoke`, { method: 'PUT' }),
  delete: (id: string) => request<void>(`/keys/platform/${id}`, { method: 'DELETE' }),
}
