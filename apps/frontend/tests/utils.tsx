import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { render, renderHook, type RenderOptions } from '@testing-library/react'
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
    companyName: '测试公司',
    companyType: 'selfhosted',
    authzRevision: 0,
    user: { name: '管理员' },
    member: mockMember,
    permissions,
    readOnly,
    billingCurrency: 'CNY',
    quotaPerUnit: 500000,
  }
}

// ponytail: DeepPartial 让测试只需 mock 用到的方法，withOverrides 负责和 base 合并。
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K]
}

type ApiNamespaceOverrides = {
  [K in keyof AppApis]?: DeepPartial<AppApis[K]>
}

function withOverrides<K extends keyof AppApis>(
  base: AppApis,
  key: K,
  partial?: DeepPartial<AppApis[K]>,
): AppApis[K] {
  if (!partial) return base[key]
  // For nested API objects (like orgApi, keysApi), merge one level deeper
  if (key === 'orgApi' && partial) {
    const baseOrg = base.orgApi
    const partialOrg = partial as DeepPartial<typeof baseOrg>
    return {
      dataSource: { ...baseOrg.dataSource, ...partialOrg.dataSource },
      sync: { ...baseOrg.sync, ...partialOrg.sync },
      departments: { ...baseOrg.departments, ...partialOrg.departments },
      members: { ...baseOrg.members, ...partialOrg.members },
      roles: { ...baseOrg.roles, ...partialOrg.roles },
    } as AppApis[K]
  }
  if (key === 'keysApi' && partial) {
    const baseKeys = base.keysApi
    const partialKeys = partial as DeepPartial<typeof baseKeys>
    return {
      provider: { ...baseKeys.provider, ...partialKeys.provider },
      platform: { ...baseKeys.platform, ...partialKeys.platform },
    } as AppApis[K]
  }
  if (key === 'modelsApi' && partial) {
    const baseModels = base.modelsApi
    const partialModels = partial as DeepPartial<typeof baseModels>
    return {
      ...baseModels,
      ...partialModels,
      routing: { ...baseModels.routing, ...partialModels.routing },
    } as AppApis[K]
  }
  return { ...base[key], ...partial } as AppApis[K]
}

function mergeApis(base: AppApis, overrides: ApiNamespaceOverrides): AppApis {
  return {
    authApi: withOverrides(base, 'authApi', overrides.authApi),
    billingApi: withOverrides(base, 'billingApi', overrides.billingApi),
    budgetApi: withOverrides(base, 'budgetApi', overrides.budgetApi),
    auditApi: withOverrides(base, 'auditApi', overrides.auditApi),
    dashboardApi: withOverrides(base, 'dashboardApi', overrides.dashboardApi),
    modelsApi: withOverrides(base, 'modelsApi', overrides.modelsApi),
    orgApi: withOverrides(base, 'orgApi', overrides.orgApi),
    keysApi: withOverrides(base, 'keysApi', overrides.keysApi),
    approvalApi: withOverrides(base, 'approvalApi', overrides.approvalApi),
    meApi: withOverrides(base, 'meApi', overrides.meApi),
    notificationApi: withOverrides(base, 'notificationApi', overrides.notificationApi),
    sessionApi: withOverrides(base, 'sessionApi', overrides.sessionApi),
    platformApi: withOverrides(base, 'platformApi', overrides.platformApi),
  }
}

export function createMockApis(overrides: ApiNamespaceOverrides = {}): AppApis {
  const session = createMockSession()
  const base: AppApis = {
    ...defaultApis,
    orgApi: {
      ...defaultApis.orgApi,
      departments: {
        ...defaultApis.orgApi.departments,
        getTree: vi.fn().mockResolvedValue(mockDepartments),
      },
      members: {
        ...defaultApis.orgApi.members,
        list: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 }),
      },
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

/**
 * Creates a synchronous test wrapper with all providers EXCEPT the router.
 * Use this for components/hooks that don't depend on routing.
 * For routing-dependent tests, create a dedicated router in the test file.
 */
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
