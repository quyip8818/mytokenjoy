import { useCallback } from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'

export function useDeptSelection() {
  const search = useRouterState({ select: (s) => s.location.search }) as { dept?: string }
  const navigate = useNavigate()

  const selectedDeptId = search.dept ?? null

  const setSelectedDeptId = useCallback(
    (deptId: string | null) => {
      void navigate({
        to: '.',
        search: (prev: Record<string, string | undefined>) => ({
          ...prev,
          dept: deptId || undefined,
        }),
        replace: true,
      })
    },
    [navigate],
  )

  return { selectedDeptId, setSelectedDeptId }
}
