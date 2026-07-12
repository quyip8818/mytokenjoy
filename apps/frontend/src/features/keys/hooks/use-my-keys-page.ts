import { useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { PlatformKey } from '@/api/types'
import { useSession } from '@/features/session'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import { BUDGET_INSUFFICIENT_MESSAGE } from '../lib/constants'

export function useMyKeysPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { memberId } = useSession()
  const applyBudgetCta = useCtaHighlight('APPLY_BUDGET')
  const createKeyCta = useCtaHighlight('CREATE_KEY')
  const { flashRow, rowClass } = useRowHighlight()
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
    data: budgetSummary = null,
    loading: budgetLoading,
    error: budgetError,
    refresh: refreshBudget,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.keys.budget(memberId),
    queryFn: (apis) => apis.platformKeyApi.getBudgetSummary(memberId),
    enabled: Boolean(memberId),
  })

  const loading = keysLoading || budgetLoading
  const error = keysError ?? budgetError
  const refresh = async () => {
    await Promise.all([refreshKeys(), refreshBudget()])
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

  const handleToggleWithFlash = async (key: PlatformKey) => {
    const id = await handleToggle(key)
    flashRow(id)
  }

  const openCreateKey = (options?: { name?: string; budget?: string }) => {
    if (budgetSummary !== null && budgetSummary.remaining <= 0) {
      toast.error(BUDGET_INSUFFICIENT_MESSAGE)
      return
    }
    openWithRefresh('key-create', {
      initialName: options?.name,
      initialBudget: options?.budget,
    })
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
    budgetSummary,
    loading,
    error,
    deleteTarget,
    setDeleteTarget,
    applyBudgetCta,
    createKeyCta,
    handleDelete,
    handleToggle,
    handleToggleWithFlash,
    rowClass,
    openCreateKey,
    openEditKey,
    openRotateKey,
    openWithRefresh,
    refresh,
  }
}
