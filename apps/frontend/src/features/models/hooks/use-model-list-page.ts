import { useCallback, useMemo, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { ModelInfo } from '@/api/types'
import { apiErrorMessage } from '@/lib/api-error-toast'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { usePermissions } from '@/features/session'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import { useSession } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { isCustomModel, matchesModelListTab } from '../lib/model-kind'

export type ModelListTab = 'all' | 'custom' | 'builtin'

export function useModelListPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useCtaHighlight('MODEL')
  const { has } = usePermissions()
  const canManage = has(PERMISSION.MODEL_MANAGE)
  const session = useSession()
  const isSelfHosted = session.companyType === 'selfhosted'
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

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.models.all],
    flashRow,
  })

  const filteredModels = useMemo(() => {
    if (!isSelfHosted) {
      // SaaS: only show builtin models
      return models.filter((model) => !isCustomModel(model))
    }
    return models.filter((model) => matchesModelListTab(model, tab))
  }, [models, tab, isSelfHosted])

  const counts = useMemo(
    () => ({
      all: models.length,
      custom: models.filter((model) => isCustomModel(model)).length,
      builtin: models.filter((model) => !isCustomModel(model)).length,
    }),
    [models],
  )

  const toggleMutation = useMutation({
    mutationFn: (model: ModelInfo) => apis.modelsApi.toggle(model.modelId, !model.active),
    onSuccess: (_data, model) => {
      toast.success(model.active ? '模型已禁用' : '模型已启用')
      flashRow(model.modelId)
      void refresh()
    },
    onError: (err) => toast.error(apiErrorMessage(err, '操作失败')),
  })

  const deleteMutation = useMutation({
    mutationFn: (model: ModelInfo) => apis.modelsApi.delete(model.modelId),
    onSuccess: (_data, model) => {
      toast.success('模型已删除')
      flashRow(model.modelId)
      void refresh()
    },
    onError: (err) => toast.error(apiErrorMessage(err, '删除失败')),
  })

  const handleToggle = useCallback(
    (model: ModelInfo) => toggleMutation.mutateAsync(model).then(() => {}).catch(() => {}),
    [toggleMutation],
  )

  const handleDelete = useCallback(
    (model: ModelInfo) => { deleteMutation.mutate(model) },
    [deleteMutation],
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
    isSelfHosted,
    modelCta,
    rowClass,
    handleToggle,
    handleDelete,
    mutating: toggleMutation.isPending || deleteMutation.isPending,
    openCreate,
    openEdit,
  }
}
