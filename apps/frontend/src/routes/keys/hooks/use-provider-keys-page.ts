import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { ProviderKey } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function useProviderKeysPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { flashRow, rowClass } = useRowHighlight()
  const {
    data: keys = [],
    loading,
    error,
    refresh,
  } = useAsyncResource(() => apis.providerKeyApi.list(), [apis])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

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
