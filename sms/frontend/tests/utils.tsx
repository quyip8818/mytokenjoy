import type { ReactNode } from 'react'
import { vi } from 'vitest'
import { render, renderHook, type RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider } from '@/features/query'

type ApiNamespaceOverrides = {
  [K in keyof AppApis]?: Partial<AppApis[K]>
}

function mergeApis(base: AppApis, overrides: ApiNamespaceOverrides): AppApis {
  const result = { ...base }
  for (const key of Object.keys(overrides) as (keyof AppApis)[]) {
    result[key] = { ...base[key], ...overrides[key] } as AppApis[typeof key]
  }
  return result
}

export function createMockApis(overrides: ApiNamespaceOverrides = {}): AppApis {
  const base: AppApis = {
    ...defaultApis,
    authApi: {
      ...defaultApis.authApi,
      login: vi.fn().mockResolvedValue({
        accessToken: 'tok',
        user: { id: 'u1', username: 'admin', realName: '管理员', role: 'admin' },
      }),
      logout: vi.fn().mockResolvedValue(undefined),
    },
  }
  return mergeApis(base, overrides)
}

export interface TestWrapperOptions {
  apis?: AppApis
  initialEntries?: string[]
}

export function createTestWrapper(options: TestWrapperOptions = {}) {
  const apis = options.apis ?? createMockApis()

  return function TestWrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={options.initialEntries ?? ['/']}>
        <QueryProvider>
          <ApiProvider apis={apis}>{children}</ApiProvider>
        </QueryProvider>
      </MemoryRouter>
    )
  }
}

export function renderWithProviders(ui: ReactNode, options?: TestWrapperOptions & RenderOptions) {
  const { initialEntries, apis, ...renderOptions } = options ?? {}
  return render(ui, {
    wrapper: createTestWrapper({ initialEntries, apis }),
    ...renderOptions,
  })
}

export function renderHookWithProviders<TResult>(
  hook: () => TResult,
  options?: TestWrapperOptions,
) {
  return renderHook(hook, { wrapper: createTestWrapper(options) })
}
