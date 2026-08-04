import { useCallback, useMemo, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { withErrorToast } from '@/lib/api-error-toast'
import { mapProjectsToViews } from '../lib/mappers'
import { alertRuleToView, alertRuleFromView, type AlertRuleView } from '../lib/alerts'
import type { AlertTypeFilter, AlertStatusFilter } from '../components/budget-alerts-toolbar'
import type { AlertStats } from '../components/budget-alerts-stats'

export function useBudgetAlertRulesPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRuleView | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AlertRuleView | null>(null)

  // Filter state
  const [typeFilter, setTypeFilter] = useState<AlertTypeFilter>('all')
  const [statusFilter, setStatusFilter] = useState<AlertStatusFilter>('all')
  const [search, setSearch] = useState('')

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
    queryFn: (api) => api.orgApi.roles.list(),
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

  // Stats computation
  const stats: AlertStats = useMemo(() => {
    const totalTeams = nodeNameMap.size
    const totalProjects = projectsData.length
    const coveredTeams = new Set(
      ruleViews.filter((r) => r.targetType === 'team').map((r) => r.targetId),
    ).size
    const coveredProjects = new Set(
      ruleViews.filter((r) => r.targetType === 'project').map((r) => r.targetId),
    ).size

    return {
      total: ruleViews.length,
      enabled: ruleViews.filter((r) => r.enabled).length,
      teamCoverage: { covered: coveredTeams, total: totalTeams },
      projectCoverage: { covered: coveredProjects, total: totalProjects },
    }
  }, [ruleViews, nodeNameMap, projectsData])

  // Filtered rules
  const filteredRules = useMemo(() => {
    let result = ruleViews
    if (typeFilter !== 'all') {
      result = result.filter((r) => r.targetType === typeFilter)
    }
    if (statusFilter !== 'all') {
      result = result.filter((r) => (statusFilter === 'enabled' ? r.enabled : !r.enabled))
    }
    if (search.trim()) {
      const keyword = search.trim().toLowerCase()
      result = result.filter((r) => r.targetName.toLowerCase().includes(keyword))
    }
    return result
  }, [ruleViews, typeFilter, statusFilter, search])

  const handleToggle = useCallback(
    async (rule: AlertRuleView) => {
      await withErrorToast(async () => {
        await apis.budgetApi.updateAlert(rule.id, { enabled: !rule.enabled })
        toast.success(rule.enabled ? '已禁用' : '已启用')
        await refresh()
      }, '操作失败')
    },
    [apis, refresh],
  )

  const handleDelete = useCallback(async () => {
    if (!deleteTarget) return
    await withErrorToast(async () => {
      await apis.budgetApi.deleteAlert(deleteTarget.id)
      setDeleteTarget(null)
      toast.success('已删除')
      await refresh()
    }, '删除失败')
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
      await withErrorToast(async () => {
        const payload = alertRuleFromView(view)
        if (existingId) {
          await apis.budgetApi.updateAlert(existingId, payload)
        } else {
          await apis.budgetApi.createAlert(payload)
        }
        toast.success(existingId ? '规则已更新' : '规则已创建')
        await refresh()
      }, '保存失败')
    },
    [apis, refresh],
  )

  return {
    rules: filteredRules,
    allRules: ruleViews,
    projects,
    tree,
    roles,
    stats,
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
    typeFilter,
    setTypeFilter,
    statusFilter,
    setStatusFilter,
    search,
    setSearch,
  }
}
