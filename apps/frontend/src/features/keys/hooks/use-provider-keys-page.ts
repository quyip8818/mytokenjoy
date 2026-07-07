import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { ProviderKey } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'

export function useProviderKeysPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const {
    data: keys = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.keys.provider(),
    queryFn: (a) => a.providerKeyApi.list(),
  })
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
    flashRow,
  })

  const handleToggle = useCallback(
    async (key: ProviderKey) => {
      const enabled = key.status !== 'active'
      await apis.providerKeyApi.toggle(key.id, enabled)
      toast.success(enabled ? 'Key 已启用' : 'Key 已禁用')
      flashRow(key.id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

  const handleDelete = useCallback(
    async (id: string) => {
      await apis.providerKeyApi.delete(id)
      toast.success('Key 已删除')
      void refresh()
    },
    [apis, refresh],
  )

  const openForm = useCallback(() => openWithRefresh('provider-key-form'), [openWithRefresh])

  return {
    keys,
    loading,
    error,
    refresh,
    rowClass,
    handleToggle,
    handleDelete,
    openForm,
  }
}
