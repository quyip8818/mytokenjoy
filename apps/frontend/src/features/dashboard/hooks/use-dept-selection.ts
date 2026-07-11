import { useCallback } from 'react'
import { useSearchParams } from 'react-router'

const DEPT_PARAM = 'dept'

export function useDeptSelection() {
  const [searchParams, setSearchParams] = useSearchParams()

  const selectedDeptId = searchParams.get(DEPT_PARAM)

  const setSelectedDeptId = useCallback(
    (deptId: string | null) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          if (deptId) {
            next.set(DEPT_PARAM, deptId)
          } else {
            next.delete(DEPT_PARAM)
          }
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )

  return { selectedDeptId, setSelectedDeptId }
}
