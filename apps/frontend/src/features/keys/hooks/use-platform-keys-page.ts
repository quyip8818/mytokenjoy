import { useCallback, useMemo, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { Department } from '@/api/types'
import { filterDepartmentTree } from '@/features/org'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import type { PlatformKeyTab } from '../lib/types'

export type { PlatformKeyTab } from '../lib/types'

function collectExpandedIds(departments: Department[], ids = new Set<string>()): Set<string> {
  for (const dept of departments) {
    ids.add(dept.id)
    if (dept.children) collectExpandedIds(dept.children, ids)
  }
  return ids
}

export function usePlatformKeysPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const [selectedDeptId, setSelectedDeptId] = useState<string | undefined>()
  const [activeTab, setActiveTab] = useState<PlatformKeyTab>('member')
  const [treeSearch, setTreeSearch] = useState('')
  const [search, setSearch] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  const {
    data: departments = [],
    loading: treeLoading,
    error: treeError,
    refresh: refreshTree,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.org.departmentTree(),
    queryFn: (a) => a.departmentApi.getTree(),
  })

  const {
    data: keys = [],
    loading: keysLoading,
    error: keysError,
    refresh: refreshKeys,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.keys.platformList(selectedDeptId, activeTab),
    queryFn: (a) =>
      a.platformKeyApi
        .list({
          departmentId: selectedDeptId,
          scope: activeTab,
        })
        .then((res) => res.items),
  })

  const { openWithRefresh } = useWorkflowRefresh({
    refresh: refreshKeys,
    invalidateKeys: [queryKeys.keys.all],
    flashRow,
  })

  const filteredTree = useMemo(
    () => filterDepartmentTree(departments, treeSearch.trim()),
    [departments, treeSearch],
  )

  const effectiveExpanded = useMemo(() => {
    if (!treeSearch.trim()) return expanded
    return collectExpandedIds(filteredTree)
  }, [expanded, filteredTree, treeSearch])

  const filteredKeys = useMemo(() => {
    if (!search.trim()) return keys
    const lower = search.toLowerCase()
    return keys.filter((key) => {
      if (activeTab === 'member') {
        return (
          key.name.toLowerCase().includes(lower) ||
          (key.memberName?.toLowerCase().includes(lower) ?? false) ||
          key.keyPrefix.toLowerCase().includes(lower)
        )
      }
      if (activeTab === 'project_member') {
        return (
          key.name.toLowerCase().includes(lower) ||
          (key.memberName?.toLowerCase().includes(lower) ?? false) ||
          (key.projectName?.toLowerCase().includes(lower) ?? false) ||
          key.keyPrefix.toLowerCase().includes(lower)
        )
      }
      return (
        key.name.toLowerCase().includes(lower) ||
        (key.projectName?.toLowerCase().includes(lower) ?? false) ||
        key.keyPrefix.toLowerCase().includes(lower)
      )
    })
  }, [activeTab, keys, search])

  const refresh = useCallback(async () => {
    await Promise.all([refreshTree(), refreshKeys()])
  }, [refreshKeys, refreshTree])

  const toggleExpand = useCallback((id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const handleRevoke = useCallback(
    async (id: string) => {
      await apis.platformKeyApi.revoke(id)
      toast.success('Key 已吊销')
      flashRow(id)
      void refreshKeys()
    },
    [apis, flashRow, refreshKeys],
  )

  const openCreateKey = useCallback(
    () => openWithRefresh('key-create', { adminCreate: true, scope: activeTab }),
    [activeTab, openWithRefresh],
  )

  return {
    departments: filteredTree,
    selectedDeptId,
    setSelectedDeptId,
    activeTab,
    setActiveTab,
    treeSearch,
    setTreeSearch,
    search,
    setSearch,
    expanded: effectiveExpanded,
    toggleExpand,
    keys: filteredKeys,
    loading: treeLoading || keysLoading,
    error: treeError || keysError,
    refresh,
    rowClass,
    handleRevoke,
    openCreateKey,
  }
}
