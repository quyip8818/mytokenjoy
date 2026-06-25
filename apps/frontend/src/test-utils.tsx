import { ROUTES } from '@/config/routes'
import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { render, renderHook, type RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import type { Department, Member } from '@/api/types'
import type { SessionContext } from '@/api/types'
import type { PermissionKey } from '@/lib/permission-keys'
import { ALL_PERMISSIONS } from '@/lib/permissions'
import { DemoProvider } from '@/features/demo'
import { createDemoRoleStore } from '@/features/demo/roles/store'
import { DEFAULT_DEMO_MEMBER_ID } from '@/features/demo/roles/constants'
import { WorkflowProvider } from '@/features/workflow/workflow-context'

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

export const mockDepartments: Department[] = [
  {
    id: 'd1',
    name: '总部',
    parentId: null,
    memberCount: 2,
    children: [
      {
        id: 'd2',
        name: '研发部',
        parentId: 'd1',
        memberCount: 1,
        children: [],
      },
    ],
  },
]

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

export function createMockApis(overrides: Partial<AppApis> = {}): AppApis {
  const session = createMockSession()
  return {
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
    ...overrides,
  }
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
        ...defaultApis.sessionApi,
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
