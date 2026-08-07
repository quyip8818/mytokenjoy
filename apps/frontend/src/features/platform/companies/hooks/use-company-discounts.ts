import { useCallback, useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { DiscountEntry, SetDiscountInput } from '@/api/types'
import { platformCompaniesKeys } from '../query-keys'

function discountKeys(companyId: string) {
  return [...platformCompaniesKeys.all, 'discounts', companyId] as const
}

export function useCompanyDiscounts(companyId: string | null) {
  const apis = useInjectedApis()

  const {
    data: discounts = [],
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    queryKey: discountKeys(companyId ?? ''),
    queryFn: (a) => a.platformApi.listCompanyDiscounts(companyId!),
    enabled: !!companyId,
  })

  const [submitting, setSubmitting] = useState(false)

  const submit = useCallback(
    async (data: SetDiscountInput) => {
      if (!companyId) return
      setSubmitting(true)
      try {
        await apis.platformApi.setCompanyDiscount(companyId, data)
        toast.success('优惠已保存')
        void refresh()
      } catch (e: unknown) {
        toast.error(e instanceof Error ? e.message : '保存失败')
      } finally {
        setSubmitting(false)
      }
    },
    [apis, companyId, refresh],
  )

  return { discounts, loading, error, refresh, submit, submitting }
}

export type UseCompanyDiscountsReturn = ReturnType<typeof useCompanyDiscounts>
