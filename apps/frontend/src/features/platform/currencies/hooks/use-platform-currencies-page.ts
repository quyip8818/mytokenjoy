import { useCallback, useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { PlatformCurrency } from '@/api/types'
import { platformCurrenciesKeys } from '../query-keys'

export function usePlatformCurrenciesPage() {
  const apis = useInjectedApis()

  const {
    data: currencies = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    queryKey: platformCurrenciesKeys.list(),
    queryFn: (a) => a.platformApi.listCurrencies(),
  })

  // --- Create dialog ---
  const [showCreate, setShowCreate] = useState(false)
  const [createCode, setCreateCode] = useState('')
  const [createQpu, setCreateQpu] = useState('')
  const [creating, setCreating] = useState(false)

  const openCreate = useCallback(() => {
    setCreateCode('')
    setCreateQpu('')
    setShowCreate(true)
  }, [])

  const closeCreate = useCallback(() => setShowCreate(false), [])

  const handleCreate = useCallback(async () => {
    const code = createCode.trim().toUpperCase()
    if (!/^[A-Z]{3}$/.test(code)) {
      toast.error('币种代码必须是 3 位大写字母')
      return
    }
    const quotaPerUnit = Number(createQpu)
    if (!quotaPerUnit || quotaPerUnit <= 0) {
      toast.error('Quota/单位必须为正整数')
      return
    }
    setCreating(true)
    try {
      await apis.platformApi.createCurrency({ code, quotaPerUnit })
      toast.success(`币种 ${code} 创建成功`)
      setShowCreate(false)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '创建失败')
    } finally {
      setCreating(false)
    }
  }, [apis, createCode, createQpu, refresh])

  // --- Edit QPU dialog ---
  const [editTarget, setEditTarget] = useState<PlatformCurrency | null>(null)
  const [editQpu, setEditQpu] = useState('')
  const [editing, setEditing] = useState(false)

  const openEdit = useCallback((c: PlatformCurrency) => {
    setEditTarget(c)
    setEditQpu(String(c.quotaPerUnit))
  }, [])

  const closeEdit = useCallback(() => setEditTarget(null), [])

  const handleEdit = useCallback(async () => {
    if (!editTarget) return
    const quotaPerUnit = Number(editQpu)
    if (!quotaPerUnit || quotaPerUnit <= 0) {
      toast.error('Quota/单位必须为正整数')
      return
    }
    setEditing(true)
    try {
      await apis.platformApi.updateCurrency(editTarget.code, { quotaPerUnit })
      toast.success(`${editTarget.code} 已更新`)
      setEditTarget(null)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '更新失败')
    } finally {
      setEditing(false)
    }
  }, [apis, editTarget, editQpu, refresh])

  // --- Toggle status ---
  const handleToggleStatus = useCallback(
    async (c: PlatformCurrency) => {
      const newEnabled = !c.enabled
      if (!newEnabled) {
        // 禁用时确认
        if (!window.confirm('该币种将无法用于新充值，确认禁用？')) return
      }
      try {
        await apis.platformApi.toggleCurrencyStatus(c.code, newEnabled)
        toast.success(newEnabled ? `${c.code} 已启用` : `${c.code} 已禁用`)
        void refresh()
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : '操作失败')
      }
    },
    [apis, refresh],
  )

  return {
    currencies,
    loading,
    error,
    refresh,
    // create
    showCreate,
    createCode,
    setCreateCode,
    createQpu,
    setCreateQpu,
    creating,
    openCreate,
    closeCreate,
    handleCreate,
    // edit
    editTarget,
    editQpu,
    setEditQpu,
    editing,
    openEdit,
    closeEdit,
    handleEdit,
    // toggle
    handleToggleStatus,
  }
}
