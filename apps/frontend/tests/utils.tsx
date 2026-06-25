import { ROUTES } from '@/config/routes'
import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { render, renderHook, type RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import type { Member } from '@/api/types'
import type { SessionContext } from '@/api/types'
import type { PermissionKey } from '@/lib/permission-keys'
import { ALL_PERMISSIONS } from '@/lib/permissions'
import { DemoProvider } from '@/features/demo'
import { createDemoRoleStore } from '@/features/demo/roles/store'
import { DEFAULT_DEMO_MEMBER_ID } from '@/features/demo/roles/constants'
import { WorkflowProvider } from '@/features/workflow/workflow-context'
import { mockDepartments } from '@tests/fixtures/departments'

export { mockDepartments }

type ApiNamespaceOverrides = {
  [K in keyof AppApis]?: Partial<AppApis[K]>
}

const mockMember: Member = {
  id: DEFAULT_DEMO_MEMBER_ID,
  name: '管理员',
  phone: '13800000000',
  email: 'admin@test.com',
  departmentId: 'd1',
  departmentName: '总部',
  status: 'active',
  roles: ['超级管理员'],
  source: 'manual',
}

export function createMockSession(
  permissions: PermissionKey[] = ALL_PERMISSIONS,
  readOnly = false,
): SessionContext {
  return {
    member: mockMember,
    permissions,
    readOnly,
  }
}

function withOverrides<K extends keyof AppApis>(
  base: AppApis,
  key: K,
  partial?: Partial<AppApis[K]>,
): AppApis[K] {
  return partial ? { ...base[key], ...partial } : base[key]
}

function mergeApis(base: AppApis, overrides: ApiNamespaceOverrides): AppApis {
  return {
    budgetApi: withOverrides(base, 'budgetApi', overrides.budgetApi),
    auditApi: withOverrides(base, 'auditApi', overrides.auditApi),
    dashboardApi: withOverrides(base, 'dashboardApi', overrides.dashboardApi),
    modelApi: withOverrides(base, 'modelApi', overrides.modelApi),
    routingApi: withOverrides(base, 'routingApi', overrides.routingApi),
    dataSourceApi: withOverrides(base, 'dataSourceApi', overrides.dataSourceApi),
    syncApi: withOverrides(base, 'syncApi', overrides.syncApi),
    departmentApi: withOverrides(base, 'departmentApi', overrides.departmentApi),
    memberApi: withOverrides(base, 'memberApi', overrides.memberApi),
    roleApi: withOverrides(base, 'roleApi', overrides.roleApi),
    providerKeyApi: withOverrides(base, 'providerKeyApi', overrides.providerKeyApi),
    platformKeyApi: withOverrides(base, 'platformKeyApi', overrides.platformKeyApi),
    approvalApi: withOverrides(base, 'approvalApi', overrides.approvalApi),
    sessionApi: withOverrides(base, 'sessionApi', overrides.sessionApi),
  }
}

export function createMockApis(overrides: ApiNamespaceOverrides = {}): AppApis {
  const session = createMockSession()
  const base: AppApis = {
    ...defaultApis,
    departmentApi: {
      ...defaultApis.departmentApi,
      getTree: vi.fn().mockResolvedValue(mockDepartments),
    },
    memberApi: {
      ...defaultApis.memberApi,
      list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
    },
    approvalApi: {
      ...defaultApis.approvalApi,
      list: vi.fn().mockResolvedValue([]),
    },
    sessionApi: {
      ...defaultApis.sessionApi,
      get: vi.fn().mockResolvedValue(session),
    },
  }
  return mergeApis(base, overrides)
}

export interface TestWrapperOptions {
  apis?: AppApis
  permissions?: PermissionKey[]
  readOnly?: boolean
  initialEntries?: string[]
}

export function createTestWrapper(options: TestWrapperOptions = {}) {
  const permissions = options.permissions ?? ALL_PERMISSIONS
  const readOnly = options.readOnly ?? false
  const apis =
    options.apis ??
    createMockApis({
      sessionApi: {
        get: vi.fn().mockResolvedValue(createMockSession(permissions, readOnly)),
      },
    })
  const roleStore = createDemoRoleStore(DEFAULT_DEMO_MEMBER_ID, apis)

  return function TestWrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={options.initialEntries ?? [ROUTES.orgStructure]}>
        <ApiProvider apis={apis}>
          <DemoProvider roleStore={roleStore}>
            <WorkflowProvider>{children}</WorkflowProvider>
          </DemoProvider>
        </ApiProvider>
      </MemoryRouter>
    )
  }
}

export function renderWithProviders(ui: ReactNode, options?: TestWrapperOptions & RenderOptions) {
  const { initialEntries, permissions, readOnly, apis, ...renderOptions } = options ?? {}
  return render(ui, {
    wrapper: createTestWrapper({ initialEntries, permissions, readOnly, apis }),
    ...renderOptions,
  })
}

export function renderHookWithProviders<TResult>(
  hook: () => TResult,
  options?: TestWrapperOptions,
) {
  return renderHook(hook, { wrapper: createTestWrapper(options) })
}
