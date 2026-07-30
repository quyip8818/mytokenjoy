import { request, buildQuery } from './client'
import type { Paginated, PlatformKey, PlatformKeyScope, ProviderKey } from './types'

export const keysApi = {
  provider: {
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
  },

  platform: {
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
    rotate: (id: string) =>
      request<PlatformKey>(`/keys/platform/${id}/rotate`, { method: 'POST' }),
    delete: (id: string) => request<void>(`/keys/platform/${id}`, { method: 'DELETE' }),
    simulateBearer: (id: string) =>
      request<{ bearer: string }>(`/keys/platform/${id}/simulate-bearer`),
  },
}
