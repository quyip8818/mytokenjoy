import { useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { PlatformKey } from '@/api/types'
import { useSession } from '@/features/session'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { QUOTA_INSUFFICIENT_MESSAGE } from '@/features/workflow/constants'

export function useMyKeysPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { memberId } = useSession()
  const applyQuotaCta = useCtaHighlight('APPLY_QUOTA')
  const createKeyCta = useCtaHighlight('CREATE_KEY')
  const [deleteTarget, setDeleteTarget] = useState<PlatformKey | null>(null)

  const {
    data: keys = [],
    loading: keysLoading,
    error: keysError,
    refresh: refreshKeys,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.keys.mine(memberId),
    queryFn: (apis) => apis.platformKeyApi.list({ memberId }).then((res) => res.items),
    enabled: Boolean(memberId),
  })
  const {
    data: quota = null,
    loading: quotaLoading,
    error: quotaError,
    refresh: refreshQuota,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.keys.quota(memberId),
    queryFn: (apis) => apis.platformKeyApi.getQuotaSummary(memberId),
    enabled: Boolean(memberId),
  })

  const loading = keysLoading || quotaLoading
  const error = keysError ?? quotaError
  const refresh = async () => {
    await Promise.all([refreshKeys(), refreshQuota()])
  }
  const { openWithRefresh, open } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
  })

  const handleDelete = async () => {
    if (!deleteTarget) return
    await apis.platformKeyApi.delete(deleteTarget.id)
    toast.success('Key 已删除')
    setDeleteTarget(null)
    refresh()
  }

  const handleToggle = async (key: PlatformKey) => {
    const enabled = key.status !== 'active'
    await apis.platformKeyApi.toggle(key.id, enabled)
    toast.success(enabled ? 'Key 已启用' : 'Key 已禁用')
    refresh()
    return key.id
  }

  const openCreateKey = () => {
    if (quota !== null && quota.remaining <= 0) {
      toast.error(QUOTA_INSUFFICIENT_MESSAGE)
      return
    }
    openWithRefresh('key-create')
  }

  const openEditKey = (key: PlatformKey) => {
    open('key-edit', { key, onSuccess: refresh })
  }

  const openRotateKey = (key: PlatformKey) => {
    open('key-rotate-confirm', {
      key,
      onRotate: (k: PlatformKey) => apis.platformKeyApi.rotate(k.id),
      onDone: refresh,
    })
  }

  return {
    keys,
    quota,
    loading,
    error,
    deleteTarget,
    setDeleteTarget,
    applyQuotaCta,
    createKeyCta,
    handleDelete,
    handleToggle,
    openCreateKey,
    openEditKey,
    openRotateKey,
    openWithRefresh,
    refresh,
  }
}
