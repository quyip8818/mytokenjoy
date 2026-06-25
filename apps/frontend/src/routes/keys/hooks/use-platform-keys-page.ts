import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useKeysListPage } from '@/routes/keys/hooks/use-keys-list-page'

export function usePlatformKeysPage(injectedApis?: AppApis) {
  const { apis, keys, loading, error, refresh, flashRow, rowClass, openWithRefresh } =
    useKeysListPage(injectedApis, 'platform')

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
