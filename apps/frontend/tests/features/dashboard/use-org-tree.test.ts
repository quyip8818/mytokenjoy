import { describe, expect, it, vi } from 'vitest'
import { useOrgTree } from '@/features/dashboard/hooks/use-org-tree'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

const mockTree = [
  {
    id: 'd1',
    name: '工程部',
    parentId: null,
    memberCount: 10,
    children: [
      { id: 'd1-1', name: '后端组', parentId: 'd1', memberCount: 5, children: [] },
      { id: 'd1-2', name: '前端组', parentId: 'd1', memberCount: 5, children: [] },
    ],
  },
  { id: 'd2', name: '产品部', parentId: null, memberCount: 8, children: [] },
]

describe('useOrgTree', () => {
  it('fetches department tree on mount', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockTree),
      },
    })
    const { result } = renderHookWithProviders(() => useOrgTree(apis), { apis })
    await waitForLoaded(result, 'loading')

    expect(result.current.departments).toHaveLength(2)
    expect(result.current.departments[0]?.name).toBe('工程部')
  })

  it('builds breadcrumb path for nested dept', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockTree),
      },
    })
    const { result } = renderHookWithProviders(() => useOrgTree(apis), { apis })
    await waitForLoaded(result, 'loading')

    expect(result.current.getBreadcrumb('d1-1')).toEqual(['工程部', '后端组'])
    expect(result.current.getBreadcrumb(null)).toEqual(['全公司'])
  })
})
