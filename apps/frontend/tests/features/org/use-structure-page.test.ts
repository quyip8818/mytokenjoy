import { act } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useStructurePage } from '@/features/org/hooks/use-structure-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { mockDepartmentTree } from '@tests/fixtures/departments'
import { waitFor } from '@testing-library/react'

describe('useStructurePage', () => {
  it('loads department tree and members on mount', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockDepartmentTree),
      },
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(() => useStructurePage(apis), { apis })

    await waitFor(() => {
      expect(result.current.departments).toEqual(mockDepartmentTree)
    })
    expect(apis.departmentApi.getTree).toHaveBeenCalled()
    expect(apis.memberApi.list).toHaveBeenCalled()
  })

  it('resets page when selecting a department', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockDepartmentTree),
      },
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(() => useStructurePage(apis), { apis })

    await waitFor(() => {
      expect(result.current.departments).toHaveLength(1)
    })

    act(() => {
      result.current.setPage(3)
    })
    expect(result.current.page).toBe(3)

    const childDept = mockDepartmentTree[0].children![0]
    act(() => {
      result.current.selectDept(childDept)
    })

    expect(result.current.page).toBe(1)
    expect(result.current.selectedDept).toEqual(childDept)
  })

  it('createDept calls api and refreshes tree', async () => {
    const getTree = vi
      .fn()
      .mockResolvedValueOnce(mockDepartmentTree)
      .mockResolvedValueOnce([
        ...mockDepartmentTree,
        { id: 'd-new', name: '新部门', parentId: 'd1', memberCount: 0, children: [] },
      ])
    const create = vi.fn().mockResolvedValue(undefined)
    const apis = createMockApis({
      departmentApi: { getTree, create },
      memberApi: {
        list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
      },
    })

    const { result } = renderHookWithProviders(() => useStructurePage(apis), { apis })

    await waitFor(() => {
      expect(result.current.departments).toEqual(mockDepartmentTree)
    })

    await act(async () => {
      await result.current.createDept('新部门', 'd1')
    })

    expect(create).toHaveBeenCalledWith({ name: '新部门', parentId: 'd1' })
    await waitFor(() => {
      expect(getTree).toHaveBeenCalledTimes(2)
    })
  })
})
