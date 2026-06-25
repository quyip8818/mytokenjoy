import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function usePlatformKeysPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { flashRow, rowClass } = useRowHighlight()
  const {
    data: keys = [],
    loading,
    error,
    refresh,
  } = useAsyncResource(async () => {
    const res = await apis.platformKeyApi.list()
    return res.items
  }, [apis])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

  const handleRevoke = useCallback(
    async (id: string) => {
      await apis.platformKeyApi.revoke(id)
      toast.success('Key 已吊销')
      flashRow(id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

  const openCreateKey = useCallback(
    () => openWithRefresh('key-create', { adminCreate: true }),
    [openWithRefresh],
  )

  return {
    keys,
    loading,
    error,
    refresh,
    rowClass,
    handleRevoke,
    openCreateKey,
  }
}
