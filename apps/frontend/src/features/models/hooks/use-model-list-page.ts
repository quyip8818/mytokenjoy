import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { ModelInfo } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { usePermissions } from '@/hooks/use-permissions'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { PERMISSION } from '@/lib/permissions'

export function useModelListPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useCtaHighlight('MODEL')
  const { has } = usePermissions()
  const canManage = has(PERMISSION.MODEL_MANAGE)
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

  const handleToggle = useCallback(
    async (model: ModelInfo) => {
      await apis.modelApi.toggle(model.id, !model.enabled)
      toast.success(model.enabled ? '模型已禁用' : '模型已启用')
      flashRow(model.id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

  const openCreate = useCallback(() => openWithRefresh('model-create'), [openWithRefresh])

  return {
    models,
    loading,
    error,
    refresh,
    canManage,
    modelCta,
    rowClass,
    handleToggle,
    openCreate,
  }
}
