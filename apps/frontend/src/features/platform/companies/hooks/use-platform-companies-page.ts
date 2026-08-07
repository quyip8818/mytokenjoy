import { useCallback, useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { PlatformCompanyOverview } from '@/api/types'
import { platformCompaniesKeys } from '../query-keys'

export function usePlatformCompaniesPage() {
  const apis = useInjectedApis()

  const {
    data: companies = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    queryKey: platformCompaniesKeys.overview(),
    queryFn: (a) => a.platformApi.companiesOverview(),
  })

  // --- Recharge dialog ---
  const [rechargeTarget, setRechargeTarget] = useState<PlatformCompanyOverview | null>(null)
  const [rechargeAmount, setRechargeAmount] = useState('')
  const [recharging, setRecharging] = useState(false)

  const openRecharge = useCallback((co: PlatformCompanyOverview) => {
    setRechargeTarget(co)
    setRechargeAmount('')
  }, [])

  const closeRecharge = useCallback(() => setRechargeTarget(null), [])

  const handleRecharge = useCallback(async () => {
    if (!rechargeTarget) return
    const amount = Number(rechargeAmount)
    if (amount <= 0) {
      toast.error('请输入有效金额')
      return
    }
    setRecharging(true)
    try {
      await apis.platformApi.rechargeCompany(rechargeTarget.id, amount)
      toast.success(`已为 ${rechargeTarget.name} 充值 ¥${amount}`)
      setRechargeTarget(null)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '充值失败')
    } finally {
      setRecharging(false)
    }
  }, [apis, rechargeTarget, rechargeAmount, refresh])

  // --- Gift ---
  const [giftTarget, setGiftTarget] = useState<PlatformCompanyOverview | null>(null)
  const [giftAmount, setGiftAmount] = useState('')
  const [gifting, setGifting] = useState(false)

  const openGift = useCallback((co: PlatformCompanyOverview) => {
    setGiftTarget(co)
    setGiftAmount('')
  }, [])

  const closeGift = useCallback(() => setGiftTarget(null), [])

  const handleGift = useCallback(async () => {
    if (!giftTarget) return
    const amount = Number(giftAmount)
    if (amount <= 0) {
      toast.error('请输入有效金额')
      return
    }
    setGifting(true)
    try {
      await apis.platformApi.giftCompany(giftTarget.id, amount)
      toast.success(`已为 ${giftTarget.name} 赠送 ¥${amount}`)
      setGiftTarget(null)
      void refresh()
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : '赠送失败')
    } finally {
      setGifting(false)
    }
  }, [apis, giftTarget, giftAmount, refresh])

  // --- Toggle status ---
  const handleToggleStatus = useCallback(
    async (co: PlatformCompanyOverview) => {
      const newStatus = co.status === 'active' ? 'suspended' : 'active'
      try {
        await apis.platformApi.updateCompany(co.id, { status: newStatus })
        toast.success(newStatus === 'active' ? `${co.name} 已启用` : `${co.name} 已停用`)
        void refresh()
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : '操作失败')
      }
    },
    [apis, refresh],
  )

  // --- Discount sheet ---
  const [discountTarget, setDiscountTarget] = useState<PlatformCompanyOverview | null>(null)

  const openDiscount = useCallback((co: PlatformCompanyOverview) => {
    setDiscountTarget(co)
  }, [])

  const closeDiscount = useCallback(() => setDiscountTarget(null), [])

  return {
    companies,
    loading,
    error,
    refresh,
    // recharge
    rechargeTarget,
    rechargeAmount,
    setRechargeAmount,
    recharging,
    openRecharge,
    closeRecharge,
    handleRecharge,
    // gift
    giftTarget,
    giftAmount,
    setGiftAmount,
    gifting,
    openGift,
    closeGift,
    handleGift,
    // status
    handleToggleStatus,
    // discount
    discountTarget,
    openDiscount,
    closeDiscount,
  }
}
