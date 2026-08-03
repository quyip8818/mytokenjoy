import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { PlatformModel } from '@/api/types'
import type { ModelFormData } from '../components/model-form-dialog'
import { platformKeys } from '../query-keys'

export function usePlatformModelsPage() {
  const apis = useInjectedApis()

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

  // --- Form dialog state ---
  const [formOpen, setFormOpen] = useState(false)
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create')
  const [formModel, setFormModel] = useState<PlatformModel | null>(null)
  const [formBusy, setFormBusy] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const openCreate = useCallback(() => {
    setFormMode('create')
    setFormModel(null)
    setFormError(null)
    setFormOpen(true)
  }, [])

  const openEdit = useCallback((model: PlatformModel) => {
    setFormMode('edit')
    setFormModel(model)
    setFormError(null)
    setFormOpen(true)
  }, [])

  const handleFormSubmit = useCallback(
    async (data: ModelFormData) => {
      setFormBusy(true)
      setFormError(null)
      try {
        if (formMode === 'create') {
          await apis.platformApi.createModel({
            type: data.type,
            name: data.name,
            provider: data.provider,
            inputPrice: data.inputPrice,
            outputPrice: data.outputPrice,
            cacheInputPrice: data.cacheInputPrice,
            capabilities: data.capabilities,
            maxContext: data.maxContext,
          })
          toast.success('模型已添加')
        } else if (formModel) {
          await apis.platformApi.updateModel(formModel.modelId, {
            name: data.name,
            provider: data.provider,
            capabilities: data.capabilities,
            maxContext: data.maxContext,
          })
          // Update pricing separately if changed
          if (
            data.inputPrice !== formModel.inputPrice ||
            data.outputPrice !== formModel.outputPrice ||
            data.cacheInputPrice !== formModel.cacheInputPrice
          ) {
            await apis.platformApi.setPricing(formModel.modelId, {
              inputPrice: data.inputPrice,
              outputPrice: data.outputPrice,
              cacheInputPrice: data.cacheInputPrice,
            })
          }
          toast.success('模型已更新')
        }
        setFormOpen(false)
        void refresh()
      } catch (e: unknown) {
        setFormError(e instanceof Error ? e.message : '操作失败')
      } finally {
        setFormBusy(false)
      }
    },
    [apis, formMode, formModel, refresh],
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
    // Form dialog
    formOpen,
    setFormOpen,
    formMode,
    formModel,
    formBusy,
    formError,
    openCreate,
    openEdit,
    handleFormSubmit,
  }
}
