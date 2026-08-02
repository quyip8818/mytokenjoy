import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { PlatformModel } from '@/api/types'
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
  const [pricingModel, setPricingModel] = useState<PlatformModel | null>(null)
  const [pricingForm, setPricingForm] = useState({ inputPrice: '', outputPrice: '' })

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
        await apis.platformApi.updateModel(model.modelId, { active: !model.active })
        toast.success(model.active ? '模型已禁用' : '模型已启用')
        void refresh()
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : '操作失败')
      }
    },
    [apis, refresh],
  )

  const openPricing = useCallback((model: PlatformModel) => {
    setPricingModel(model)
    setPricingForm({
      inputPrice: model.inputPrice > 0 ? String(model.inputPrice) : '',
      outputPrice: model.outputPrice > 0 ? String(model.outputPrice) : '',
    })
  }, [])

  const handleSavePricing = useCallback(async () => {
    if (!pricingModel) return
    try {
      await apis.platformApi.setPricing(pricingModel.modelId, {
        inputPrice: Number(pricingForm.inputPrice) || 0,
        outputPrice: Number(pricingForm.outputPrice) || 0,
      })
      toast.success('定价已更新')
      setPricingModel(null)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '更新失败')
    }
  }, [apis, pricingModel, pricingForm, refresh])

  const closePricing = useCallback(() => setPricingModel(null), [])

  return {
    models,
    loading,
    error,
    refresh,
    publishing,
    handlePublish,
    handleToggle,
    pricingModel,
    pricingForm,
    setPricingForm,
    openPricing,
    closePricing,
    handleSavePricing,
  }
}
