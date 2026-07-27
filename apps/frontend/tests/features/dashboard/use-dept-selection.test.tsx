import { describe, expect, it } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { z } from 'zod'
import { useDeptSelection } from '@/features/dashboard/hooks/use-dept-selection'

function createWrapper(initialEntry: string) {
  const rootRoute = createRootRoute()
  const dashboardRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/dashboard/cost',
    validateSearch: z.object({ dept: z.string().optional() }),
  })
  const routeTree = rootRoute.addChildren([dashboardRoute])
  const history = createMemoryHistory({ initialEntries: [initialEntry] })
  const router = createRouter({ routeTree, history })

  return function Wrapper({ children }: { children: ReactNode }) {
    return <RouterProvider router={router} defaultComponent={() => <>{children}</>} />
  }
}

describe('useDeptSelection', () => {
  it('returns null when no dept param', async () => {
    const { result } = renderHook(() => useDeptSelection(), {
      wrapper: createWrapper('/dashboard/cost'),
    })
    await waitFor(() => {
      expect(result.current).not.toBeNull()
    })
    expect(result.current.selectedDeptId).toBeNull()
  })

  it('reads dept from URL search param', async () => {
    const { result } = renderHook(() => useDeptSelection(), {
      wrapper: createWrapper('/dashboard/cost?dept=d1'),
    })
    await waitFor(() => {
      expect(result.current).not.toBeNull()
    })
    expect(result.current.selectedDeptId).toBe('d1')
  })

  it('updates URL when setSelectedDeptId is called', async () => {
    const { result } = renderHook(() => useDeptSelection(), {
      wrapper: createWrapper('/dashboard/cost'),
    })
    await waitFor(() => {
      expect(result.current).not.toBeNull()
    })
    await act(async () => {
      result.current.setSelectedDeptId('d2')
    })
    await waitFor(() => {
      expect(result.current.selectedDeptId).toBe('d2')
    })
  })

  it('clears dept param when set to null', async () => {
    const { result } = renderHook(() => useDeptSelection(), {
      wrapper: createWrapper('/dashboard/cost?dept=d1'),
    })
    await waitFor(() => {
      expect(result.current).not.toBeNull()
    })
    await act(async () => {
      result.current.setSelectedDeptId(null)
    })
    await waitFor(() => {
      expect(result.current.selectedDeptId).toBeNull()
    })
  })
})
