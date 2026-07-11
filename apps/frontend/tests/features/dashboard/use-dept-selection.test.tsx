import { describe, expect, it } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { useDeptSelection } from '@/features/dashboard/hooks/use-dept-selection'

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter initialEntries={['/dashboard/cost']}>{children}</MemoryRouter>
}

function wrapperWithDept({ children }: { children: React.ReactNode }) {
  return <MemoryRouter initialEntries={['/dashboard/cost?dept=d1']}>{children}</MemoryRouter>
}

describe('useDeptSelection', () => {
  it('returns null when no dept param', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper })
    expect(result.current.selectedDeptId).toBeNull()
  })

  it('reads dept from URL search param', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper: wrapperWithDept })
    expect(result.current.selectedDeptId).toBe('d1')
  })

  it('updates URL when setSelectedDeptId is called', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper })
    act(() => {
      result.current.setSelectedDeptId('d2')
    })
    expect(result.current.selectedDeptId).toBe('d2')
  })

  it('clears dept param when set to null', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper: wrapperWithDept })
    act(() => {
      result.current.setSelectedDeptId(null)
    })
    expect(result.current.selectedDeptId).toBeNull()
  })
})
