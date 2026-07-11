import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { Department } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

function buildPathMap(departments: Department[]): Map<string, string[]> {
  const map = new Map<string, string[]>()

  function walk(nodes: Department[], path: string[]) {
    for (const node of nodes) {
      const currentPath = [...path, node.name]
      map.set(node.id, currentPath)
      if (node.children && node.children.length > 0) {
        walk(node.children, currentPath)
      }
    }
  }

  walk(departments, [])
  return map
}

export function useOrgTree(injectedApis?: AppApis) {
  const {
    data: departments = [],
    loading,
    error,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.tree(),
    queryFn: (apis) => apis.departmentApi.getTree(),
  })

  const pathMap = useMemo(() => buildPathMap(departments), [departments])

  const getBreadcrumb = useCallback(
    (deptId: string | null): string[] => {
      if (!deptId) return ['全公司']
      return pathMap.get(deptId) ?? ['全公司']
    },
    [pathMap],
  )

  return { departments, loading, error, getBreadcrumb }
}
