import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { mapProjectsToViews } from '../lib/mappers'
import { alertRuleToView, alertRuleFromView, type AlertRuleView } from '../lib/alerts'

export function useBudgetAlertRulesPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRuleView | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AlertRuleView | null>(null)

  const {
    data: rules = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.alerts(),
    queryFn: (api) => api.budgetApi.getAlerts(),
  })

  const { data: projectsData = [] } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.projects(),
    queryFn: (api) => api.budgetApi.getProjects(),
  })

  const { data: tree = [] } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.tree(),
    queryFn: (api) => api.budgetApi.getTree(),
  })

  const { data: roles = [] } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.roles(),
    queryFn: (api) => api.roleApi.list(),
  })

  const ruleViews = useMemo(
    () => rules.map((rule) => alertRuleToView(rule, projectsData)),
    [rules, projectsData],
  )

  const nodeNameMap = useMemo(() => {
    const map = new Map<string, string>()
    function walk(nodes: typeof tree) {
      for (const node of nodes) {
        map.set(node.id, node.name)
        if (node.children) walk(node.children)
      }
    }
    walk(tree)
    return map
  }, [tree])

  const projects = useMemo(
    () => mapProjectsToViews(projectsData, nodeNameMap, tree[0]?.period ?? ''),
    [projectsData, nodeNameMap, tree],
  )

  const handleToggle = useCallback(
    async (rule: AlertRuleView) => {
      await apis.budgetApi.updateAlert(rule.id, { enabled: !rule.enabled })
      await refresh()
    },
    [apis, refresh],
  )

  const handleDelete = useCallback(async () => {
    if (!deleteTarget) return
    await apis.budgetApi.deleteAlert(deleteTarget.id)
    setDeleteTarget(null)
    await refresh()
  }, [apis, deleteTarget, refresh])

  const openCreate = useCallback(() => {
    setEditingRule(null)
    setDialogOpen(true)
  }, [])

  const openEdit = useCallback((rule: AlertRuleView) => {
    setEditingRule(rule)
    setDialogOpen(true)
  }, [])

  const saveRule = useCallback(
    async (view: AlertRuleView, existingId?: string) => {
      const payload = alertRuleFromView(view)
      if (existingId) {
        await apis.budgetApi.updateAlert(existingId, payload)
      } else {
        await apis.budgetApi.createAlert(payload)
      }
      await refresh()
    },
    [apis, refresh],
  )

  return {
    rules: ruleViews,
    projects,
    tree,
    roles,
    loading,
    error,
    refresh,
    dialogOpen,
    setDialogOpen,
    editingRule,
    deleteTarget,
    setDeleteTarget,
    handleToggle,
    handleDelete,
    openCreate,
    openEdit,
    saveRule,
  }
}
