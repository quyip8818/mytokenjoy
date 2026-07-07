import { useCallback, useMemo, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { ModelInfo } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { usePermissions } from '@/hooks/use-permissions'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import { PERMISSION } from '@/lib/permissions'

export type ModelListTab = 'all' | 'custom' | 'builtin'

export function useModelListPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useCtaHighlight('MODEL')
  const { has } = usePermissions()
  const canManage = has(PERMISSION.MODEL_MANAGE)
  const [tab, setTab] = useState<ModelListTab>('all')

  const {
    data: models = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.models.list(),
    queryFn: (a) => a.modelApi.list(),
  })

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.models.all],
    flashRow,
  })

  const filteredModels = useMemo(() => {
    if (tab === 'all') return models
    return models.filter((model) => model.type === tab)
  }, [models, tab])

  const counts = useMemo(
    () => ({
      all: models.length,
      custom: models.filter((model) => model.type === 'custom').length,
      builtin: models.filter((model) => model.type === 'builtin').length,
    }),
    [models],
  )

  const handleToggle = useCallback(
    async (model: ModelInfo) => {
      await apis.modelApi.toggle(model.id, !model.enabled)
      toast.success(model.enabled ? '模型已禁用' : '模型已启用')
      flashRow(model.id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

  const handleDelete = useCallback(
    async (model: ModelInfo) => {
      await apis.modelApi.delete(model.id)
      toast.success('模型已删除')
      flashRow(model.id)
      void refresh()
    },
    [apis, flashRow, refresh],
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
    handleDelete,
    openCreate,
    openEdit,
  }
}
