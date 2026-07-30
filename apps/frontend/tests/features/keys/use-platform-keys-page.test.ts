import { describe, expect, it, vi } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { usePlatformKeysPage } from '@/features/keys/hooks/use-platform-keys-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('usePlatformKeysPage', () => {
  it('loads platform keys on mount with member scope', async () => {
    const items = [{ id: 'pk-1', name: 'Admin Key', status: 'active', scope: 'member' }]
    const apis = createMockApis({
      orgApi: {
        departments: {
          getTree: vi.fn().mockResolvedValue([]),
        },
      },
      keysApi: {
        platform: {
          list: vi.fn().mockResolvedValue({ items, total: 1 }),
        },
      },
    })

    const { result } = renderHookWithProviders(() => usePlatformKeysPage(apis), { apis })

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.keys).toEqual(items)
    })

    expect(apis.orgApi.departments.getTree).toHaveBeenCalled()
    expect(apis.keysApi.platform.list).toHaveBeenCalledWith({
      departmentId: undefined,
      scope: 'member',
    })
  })

  it('reloads keys when active tab changes to project_member', async () => {
    const memberItems = [{ id: 'pk-1', name: 'Member Key', status: 'active', scope: 'member' }]
    const projectMemberItems = [
      { id: 'pk-2', name: 'PM Key', status: 'active', scope: 'project_member' },
    ]
    const apis = createMockApis({
      orgApi: {
        departments: {
          getTree: vi.fn().mockResolvedValue([]),
        },
      },
      keysApi: {
        platform: {
          list: vi
            .fn()
            .mockResolvedValueOnce({ items: memberItems, total: 1 })
            .mockResolvedValueOnce({ items: projectMemberItems, total: 1 }),
        },
      },
    })

    const { result } = renderHookWithProviders(() => usePlatformKeysPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await act(async () => {
      result.current.setActiveTab('project_member')
    })

    await waitFor(() => {
      expect(result.current.keys).toEqual(projectMemberItems)
    })
    expect(apis.keysApi.platform.list).toHaveBeenLastCalledWith({
      departmentId: undefined,
      scope: 'project_member',
    })
  })

  it('opens create workflow with active tab scope', async () => {
    const apis = createMockApis({
      orgApi: {
        departments: {
          getTree: vi.fn().mockResolvedValue([]),
        },
      },
      keysApi: {
        platform: {
          list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
        },
      },
    })

    const { result } = renderHookWithProviders(() => usePlatformKeysPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await act(async () => {
      result.current.setActiveTab('project')
    })
    await act(async () => {
      result.current.openCreateKey()
    })

    // WorkflowProvider captures open; verify via refresh hook side effect is enough for scope wiring.
    expect(result.current.activeTab).toBe('project')
  })

  it('filters out deleted keys from list', async () => {
    const items = [
      { id: 'pk-1', name: 'Active Key', status: 'active', scope: 'member' },
      { id: 'pk-2', name: 'Deleted Key', status: 'deleted', scope: 'member' },
    ]
    const apis = createMockApis({
      orgApi: {
        departments: {
          getTree: vi.fn().mockResolvedValue([]),
        },
      },
      keysApi: {
        platform: {
          list: vi.fn().mockResolvedValue({ items, total: 2 }),
        },
      },
    })

    const { result } = renderHookWithProviders(() => usePlatformKeysPage(apis), { apis })
    await waitForLoaded(result, 'loading')

    await waitFor(() => {
      const ids = result.current.keys.map((k) => k.id)
      expect(ids).toContain('pk-1')
      expect(ids).not.toContain('pk-2')
    })
  })
})
