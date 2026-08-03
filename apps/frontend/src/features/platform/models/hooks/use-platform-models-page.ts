import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import { useWorkflow } from '@/features/workflow'
import type { PlatformModel } from '@/api/types'
import { platformKeys } from '../query-keys'

export function usePlatformModelsPage() {
  const apis = useInjectedApis()
  const { open: openWorkflow } = useWorkflow()

  const {
    data: models = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    queryKey: platformKeys.models(),
    queryFn: (a) => a.platformApi.listModels(),
  })

  const [publishing, setPublishing] = useState(false)

  const openCreate = useCallback(() => {
    openWorkflow('platform-model-create', { onSuccess: () => void refresh() })
  }, [openWorkflow, refresh])

  const openEdit = useCallback(
    (model: PlatformModel) => {
      openWorkflow('platform-model-edit', { model, onSuccess: () => void refresh() })
    },
    [openWorkflow, refresh],
  )

  const handlePublish = useCallback(async () => {
    setPublishing(true)
    try {
      const result = await apis.platformApi.publish()
      toast.success(`已发布，版本号: ${result.version}`)
    } catch (e: unknown) {
      toast.error(`发布失败：${e instanceof Error ? e.message : '未知错误'}`)
    } finally {
      setPublishing(false)
    }
  }, [apis])

  const handleToggle = useCallback(
    async (model: PlatformModel) => {
      try {
        await apis.platformApi.updateModel(model.modelId, { deprecated: !model.deprecated })
        toast.success(model.deprecated ? '模型已恢复' : '模型已下线')
        void refresh()
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : '操作失败')
      }
    },
    [apis, refresh],
  )

  return {
    models,
    loading,
    error,
    refresh,
    publishing,
    handlePublish,
    handleToggle,
    openCreate,
    openEdit,
  }
}
