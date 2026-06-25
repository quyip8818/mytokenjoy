import { useState } from 'react'
import { toast } from 'sonner'
import { platformKeyApi } from '@/api/keys'
import type { PlatformKey } from '@/api/types'
import { useDemoRole, useDemoCta } from '@/features/demo'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { QUOTA_INSUFFICIENT_MESSAGE } from '@/features/workflow/constants'

export function useMyKeysPage() {
  const { memberId } = useDemoRole()
  const applyQuotaCta = useDemoCta('APPLY_QUOTA')
  const createKeyCta = useDemoCta('CREATE_KEY')
  const [deleteTarget, setDeleteTarget] = useState<PlatformKey | null>(null)

  const { data, loading, refresh } = useAsyncResource(async () => {
    const [keyRes, quotaRes] = await Promise.all([
      platformKeyApi.list({ memberId }),
      platformKeyApi.getQuotaSummary(memberId),
    ])
    return { keys: keyRes.items, quota: quotaRes }
  }, [memberId])

  const keys = data?.keys ?? []
  const quota = data?.quota ?? null
  const { openWithRefresh, open } = useWorkflowRefresh(refresh)

  const handleDelete = async () => {
    if (!deleteTarget) return
    await platformKeyApi.delete(deleteTarget.id)
    toast.success('Key 已删除')
    setDeleteTarget(null)
    refresh()
  }

  const handleToggle = async (key: PlatformKey) => {
    const enabled = key.status !== 'active'
    await platformKeyApi.toggle(key.id, enabled)
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
      onRotate: (k: PlatformKey) => platformKeyApi.rotate(k.id),
      onDone: refresh,
    })
  }

  return {
    keys,
    quota,
    loading,
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
