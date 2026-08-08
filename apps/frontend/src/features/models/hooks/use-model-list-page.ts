import { useCallback, useMemo, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from '@/lib/toast'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { DiscountEntry, ModelInfo } from '@/api/types'
import { apiErrorMessage } from '@/lib/api-error-toast'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { usePermissions } from '@/features/session'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import { PERMISSION } from '@/lib/permissions'
import { IS_SAAS } from '@/config/app'
import { isCustomModel, matchesModelListTab } from '../lib/model-kind'

export type ModelListTab = 'all' | 'custom' | 'builtin'

export function useModelListPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useCtaHighlight('MODEL')
  const { has } = usePermissions()
  const canManage = has(PERMISSION.MODEL_MANAGE)
  const isLocalDeploy = !IS_SAAS
  const [tab, setTab] = useState<ModelListTab>('all')

  const {
    data: models = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.models.list(),
    queryFn: (a) => a.modelsApi.list(),
  })

  const { data: discounts = [] } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.billing.discounts(),
    queryFn: (a) => a.billingApi.myDiscounts(),
  })

  // ponytail: build a lookup map for O(1) discount per model; '*' = fallback
  const discountMap = useMemo(() => {
    const map = new Map<string, DiscountEntry>()
    for (const d of discounts) map.set(d.modelType, d)
    return map
  }, [discounts])

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.models.all],
    flashRow,
  })

  const filteredModels = useMemo(() => {
    if (!isLocalDeploy) {
      // SaaS: only show builtin models
      return models.filter((model) => !isCustomModel(model))
    }
    return models.filter((model) => matchesModelListTab(model, tab))
  }, [models, tab, isLocalDeploy])

  const counts = useMemo(
    () => ({
      all: models.length,
      custom: models.filter((model) => isCustomModel(model)).length,
      builtin: models.filter((model) => !isCustomModel(model)).length,
    }),
    [models],
  )

  const toggleMutation = useMutation({
    mutationFn: (model: ModelInfo) =>
      apis.modelsApi.update(model.modelId, { deprecated: !model.deprecated }),
    onSuccess: (_data, model) => {
      toast.success(model.deprecated ? '模型已恢复' : '模型已下线')
      flashRow(model.modelId)
      void refresh()
    },
    onError: (err) => toast.error(apiErrorMessage(err, '操作失败')),
  })

  const handleToggle = useCallback(
    (model: ModelInfo) =>
      toggleMutation
        .mutateAsync(model)
        .then(() => {})
        .catch(() => {}),
    [toggleMutation],
  )

  const openCreate = useCallback(() => openWithRefresh('model-create'), [openWithRefresh])

  const openEdit = useCallback(
    (model: ModelInfo) => openWithRefresh('model-edit', { model }),
    [openWithRefresh],
  )

  return {
    models: filteredModels,
    counts,
    tab,
    setTab,
    loading,
    error,
    refresh,
    canManage,
    modelCta,
    rowClass,
    handleToggle,
    mutating: toggleMutation.isPending,
    openCreate,
    openEdit,
    discountMap,
  }
}
