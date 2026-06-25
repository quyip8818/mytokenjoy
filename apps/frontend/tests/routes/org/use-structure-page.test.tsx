import { act } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { useStructurePage } from '@/routes/org/hooks/use-structure-page'
import { createMockApis, mockDepartments, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useStructurePage', () => {
  it('loads departments and members on mount', async () => {
    const apis = createMockApis()
    const { result } = renderHookWithProviders(() => useStructurePage(apis), { apis })

    await waitForLoaded(result, 'membersLoading')

    expect(apis.departmentApi.getTree).toHaveBeenCalled()
    expect(apis.memberApi.list).toHaveBeenCalled()
    expect(result.current.departments).toEqual(mockDepartments)
    expect(result.current.members).toEqual([])
    expect(result.current.total).toBe(0)
  })

  it('updates selected department when handleSelectDept is called', async () => {
    const apis = createMockApis()
    const { result } = renderHookWithProviders(() => useStructurePage(apis), { apis })

    await waitForLoaded(result, 'membersLoading')

    const dept = mockDepartments[0].children![0]

    act(() => {
      result.current.handleSelectDept(dept)
    })

    expect(result.current.selectedDept).toEqual(dept)
    expect(apis.memberApi.list).toHaveBeenCalledWith(
      expect.objectContaining({ departmentId: dept.id }),
    )
  })
})
