import { useCallback, useMemo, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import type { Department, ModelInfo, RoutingRule } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'

export function useModelRoutingPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [selectedNodeId, setSelectedNodeId] = useState<string | undefined>()
  const [saving, setSaving] = useState(false)

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.models.routing(),
    queryFn: async (a) => {
      const [rules, departments, models] = await Promise.all([
        a.routingApi.getRules(),
        a.departmentApi.getTree(),
        a.modelApi.list(),
      ])
      return { rules, departments, models }
    },
  })

  const rules: RoutingRule[] = useMemo(() => data?.rules ?? [], [data?.rules])
  const departments: Department[] = useMemo(() => data?.departments ?? [], [data?.departments])
  const models: ModelInfo[] = useMemo(
    () => (data?.models ?? []).filter((m) => m.enabled),
    [data?.models],
  )

  const resolvedSelectedId = selectedNodeId ?? departments[0]?.id

  const selectedRule = useMemo(
    () => (resolvedSelectedId ? rules.find((r) => r.nodeId === resolvedSelectedId) : undefined),
    [rules, resolvedSelectedId],
  )

  const selectedDepartment = useMemo(() => {
    if (!resolvedSelectedId) return undefined
    function find(nodes: Department[]): Department | undefined {
      for (const node of nodes) {
        if (node.id === resolvedSelectedId) return node
        if (node.children) {
          const found = find(node.children)
          if (found) return found
        }
      }
      return undefined
    }
    return find(departments)
  }, [departments, resolvedSelectedId])

  const parentRule = useMemo(() => {
    if (!selectedDepartment?.parentId) return undefined
    return rules.find((r) => r.nodeId === selectedDepartment.parentId)
  }, [rules, selectedDepartment])

  const handleSave = useCallback(
    async (input: {
      inherited: boolean
      allowedModelIds: number[]
      defaultModelId: number | null
      fallbackModelId: number | null
    }) => {
      if (!selectedRule) return
      setSaving(true)
      try {
        await apis.routingApi.updateRule(selectedRule.id, {
          inherited: input.inherited,
          allowedModelIds: input.allowedModelIds,
          defaultModelId: input.defaultModelId,
          fallbackModelId: input.fallbackModelId,
        })
        toast.success('模型配置已保存')
        await refresh()
      } catch {
        toast.error('保存失败，请重试')
      } finally {
        setSaving(false)
      }
    },
    [apis, selectedRule, refresh],
  )

  return {
    departments,
    models,
    rules,
    selectedNodeId: resolvedSelectedId,
    setSelectedNodeId,
    selectedRule,
    selectedDepartment,
    parentRule,
    loading,
    error,
    refresh,
    saving,
    handleSave,
  }
}
