import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { PlatformKey, ProviderKey } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'

export type KeysListSource = 'platform' | 'provider'

type KeysListResult<T> = {
  apis: AppApis
  keys: T[]
  loading: boolean
  error: Error | null
  refresh: () => Promise<unknown>
  flashRow: ReturnType<typeof useRowHighlight>['flashRow']
  rowClass: ReturnType<typeof useRowHighlight>['rowClass']
  openWithRefresh: ReturnType<typeof useWorkflowRefresh>['openWithRefresh']
}

async function fetchKeysBySource(apis: AppApis, source: KeysListSource) {
  if (source === 'platform') {
    const res = await apis.platformKeyApi.list()
    return res.items
  }
  return apis.providerKeyApi.list()
}

export function useKeysListPage(
  injectedApis: AppApis | undefined,
  source: 'platform',
): KeysListResult<PlatformKey>
export function useKeysListPage(
  injectedApis: AppApis | undefined,
  source: 'provider',
): KeysListResult<ProviderKey>
export function useKeysListPage(
  injectedApis: AppApis | undefined,
  source: KeysListSource,
): KeysListResult<PlatformKey | ProviderKey> {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const queryKey = source === 'platform' ? queryKeys.keys.platform() : queryKeys.keys.provider()
  const {
    data: keys = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey,
    queryFn: (a) => fetchKeysBySource(a, source),
  })
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
    flashRow,
  })

  return {
    apis,
    keys,
    loading,
    error,
    refresh,
    flashRow,
    rowClass,
    openWithRefresh,
  }
}
