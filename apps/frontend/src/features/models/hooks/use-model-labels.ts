import { useEffect, useState } from 'react'
import type { ModelInfo } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import type { AppApis } from '@/api/app-apis'
import { buildModelIndex, modelIdLabel } from '../lib/model-catalog'

export function useModelLabels(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [index, setIndex] = useState<Map<string, ModelInfo>>(new Map())

  useEffect(() => {
    let cancelled = false
    void apis.modelApi.list().then((models) => {
      if (!cancelled) setIndex(buildModelIndex(models))
    })
    return () => {
      cancelled = true
    }
  }, [apis])

  const labelFor = (modelId: string) => modelIdLabel(modelId, index)

  return { index, labelFor }
}
