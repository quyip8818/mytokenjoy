import { request } from './client'
import type {
  Credential,
  DataSourceStatus,
  Department,
  FieldMapping,
  FieldMappingConfig,
  ImportResult,
  MappingTestResult,
  Member,
  Paginated,
  Permission,
  Platform,
  Role,
  SyncConfig,
  SyncLog,
} from './types'

// 数据源
export const dataSourceApi = {
  getStatus: () => request<DataSourceStatus>('/org/data-source/status'),
  testConnection: (credential: Credential) =>
    request<{ success: boolean; message?: string }>('/org/data-source/test', {
      method: 'POST',
      body: JSON.stringify(credential),
    }),
  save: (credential: Credential) =>
    request<void>('/org/data-source', {
      method: 'PUT',
      body: JSON.stringify(credential),
    }),
  searchMember: (keyword: string) =>
    request<{ name: string; department: string; mappingOk: boolean }>(
      `/org/data-source/search?keyword=${encodeURIComponent(keyword)}`,
    ),
  import: () =>
    request<ImportResult>('/org/data-source/import', { method: 'POST' }),
  retryImport: (ids: string[]) =>
    request<ImportResult>('/org/data-source/import/retry', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    }),
  getFieldMappings: (platform: Platform) =>
    request<FieldMapping[]>(`/org/data-source/field-mappings?platform=${platform}`),
  saveFieldMappings: (config: FieldMappingConfig) =>
    request<void>('/org/data-source/field-mappings', {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
  testFieldMapping: (platform: Platform, keyword: string) =>
    request<MappingTestResult>(
      `/org/data-source/field-mappings/test?platform=${platform}&keyword=${encodeURIComponent(keyword)}`,
    ),
}

// 同步
export const syncApi = {
  getConfig: () => request<SyncConfig>('/org/sync/config'),
  saveConfig: (config: SyncConfig) =>
    request<void>('/org/sync/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    }),
  triggerSync: () =>
    request<ImportResult>('/org/sync/trigger', { method: 'POST' }),
  getLogs: (page: number, pageSize: number) =>
    request<Paginated<SyncLog>>(
      `/org/sync/logs?page=${page}&pageSize=${pageSize}`,
    ),
}

// 部门
export const departmentApi = {
  getTree: () => request<Department[]>('/org/departments/tree'),
  create: (data: { name: string; parentId: string }) =>
    request<Department>('/org/departments', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (id: string, data: { name: string }) =>
    request<Department>(`/org/departments/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  delete: (id: string) =>
    request<void>(`/org/departments/${id}`, { method: 'DELETE' }),
}

// 成员
export const memberApi = {
  list: (params: {
    departmentId?: string
    directOnly?: boolean
    page: number
    pageSize: number
    keyword?: string
  }) => {
    const qs = new URLSearchParams()
    if (params.departmentId) qs.set('departmentId', params.departmentId)
    if (params.directOnly) qs.set('directOnly', 'true')
    qs.set('page', String(params.page))
    qs.set('pageSize', String(params.pageSize))
    if (params.keyword) qs.set('keyword', params.keyword)
    return request<Paginated<Member>>(`/org/members?${qs}`)
  },
  create: (data: Omit<Member, 'id' | 'status' | 'roles' | 'source'>) =>
    request<Member>('/org/members', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (id: string, data: Partial<Member>) =>
    request<Member>(`/org/members/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  delete: (ids: string[]) =>
    request<void>('/org/members', {
      method: 'DELETE',
      body: JSON.stringify({ ids }),
    }),
  updateStatus: (ids: string[], status: 'active' | 'inactive') =>
    request<void>('/org/members/status', {
      method: 'PUT',
      body: JSON.stringify({ ids, status }),
    }),
  transferDepartment: (ids: string[], departmentId: string) =>
    request<void>('/org/members/transfer', {
      method: 'POST',
      body: JSON.stringify({ ids, departmentId }),
    }),
  invite: (data: { email?: string; phone?: string }) =>
    request<void>('/org/members/invite', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}

// 角色
export const roleApi = {
  list: () => request<Role[]>('/org/roles'),
  create: (data: { name: string; permissions: string[] }) =>
    request<Role>('/org/roles', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (id: string, data: { name: string; permissions: string[] }) =>
    request<Role>(`/org/roles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  delete: (id: string) =>
    request<void>(`/org/roles/${id}`, { method: 'DELETE' }),
  getMembers: (roleId: string) =>
    request<Member[]>(`/org/roles/${roleId}/members`),
  addMember: (roleId: string, memberId: string) =>
    request<void>(`/org/roles/${roleId}/members`, {
      method: 'POST',
      body: JSON.stringify({ memberId }),
    }),
  removeMember: (roleId: string, memberId: string) =>
    request<void>(`/org/roles/${roleId}/members/${memberId}`, {
      method: 'DELETE',
    }),
  getPermissions: () => request<Permission[]>('/org/permissions'),
}
