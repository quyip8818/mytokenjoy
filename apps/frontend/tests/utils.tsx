import { ROUTES } from '@/config/routes'
import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { render, renderHook, type RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import type { SessionContext } from '@/api/types'
import type { PermissionKey } from '@/lib/permission-keys'
import { ALL_PERMISSIONS } from '@/lib/permissions'
import { WorkflowProvider } from '@/features/workflow'
import { QueryProvider, createTestQueryClient } from '@/features/query'
import { mockDepartments } from '@tests/fixtures/departments'
import { mockMember } from '@tests/fixtures/members'
import { TestSessionProvider } from '@tests/test-session-provider'

export { mockDepartments }

export function createMockSession(
  permissions: PermissionKey[] = ALL_PERMISSIONS,
  readOnly = false,
): SessionContext {
  return {
    companyId: '00000000-0000-7000-8000-000000000002',
    companyType: 'selfhosted',
    authzRevision: 0,
    member: mockMember,
    permissions,
    readOnly,
    billingCurrency: 'CNY',
    pointsPerUnit: 1000,
  }
}

type ApiNamespaceOverrides = {
  [K in keyof AppApis]?: Partial<AppApis[K]>
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
    authApi: withOverrides(base, 'authApi', overrides.authApi),
    billingApi: withOverrides(base, 'billingApi', overrides.billingApi),
    budgetApi: withOverrides(base, 'budgetApi', overrides.budgetApi),
    auditApi: withOverrides(base, 'auditApi', overrides.auditApi),
    dashboardApi: withOverrides(base, 'dashboardApi', overrides.dashboardApi),
    devApi: withOverrides(base, 'devApi', overrides.devApi),
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
    meApi: withOverrides(base, 'meApi', overrides.meApi),
    notificationApi: withOverrides(base, 'notificationApi', overrides.notificationApi),
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
      getCurrent: vi.fn().mockResolvedValue(session),
    },
  }
  return mergeApis(base, overrides)
}

export interface TestWrapperOptions {
  apis?: AppApis
  permissions?: PermissionKey[]
  readOnly?: boolean
  companyType?: 'selfhosted' | 'standard' | 'trial' | 'demo' | 'testing'
  initialEntries?: string[]
}

export function createTestWrapper(options: TestWrapperOptions = {}) {
  const permissions = options.permissions ?? ALL_PERMISSIONS
  const readOnly = options.readOnly ?? false
  const companyType = options.companyType ?? 'selfhosted'
  const apis =
    options.apis ??
    createMockApis({
      sessionApi: {
        getCurrent: vi.fn().mockResolvedValue(createMockSession(permissions, readOnly)),
      },
    })
  const queryClient = createTestQueryClient()

  return function TestWrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={options.initialEntries ?? [ROUTES.orgStructure]}>
        <QueryProvider client={queryClient}>
          <ApiProvider apis={apis}>
            <TestSessionProvider
              permissions={permissions}
              readOnly={readOnly}
              companyType={companyType}
            >
              <WorkflowProvider>{children}</WorkflowProvider>
            </TestSessionProvider>
          </ApiProvider>
        </QueryProvider>
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
