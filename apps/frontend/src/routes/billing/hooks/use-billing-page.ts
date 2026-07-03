import type { AppApis } from '@/api/app-apis'
import { useInjectedMutation, useInjectedQuery } from '@/features/query'
import { usePermissions } from '@/hooks/use-permissions'
import { toast } from 'sonner'

const DEMO_RECHARGE_AMOUNT = 100

export function useBillingPage(injectedApis?: AppApis) {
  const { canWrite } = usePermissions()
  const walletQuery = useInjectedQuery({
    injectedApis,
    queryKey: ['billing', 'wallet'],
    queryFn: (apis) => apis.billingApi.getWallet(),
  })
  const rechargeMutation = useInjectedMutation({
    injectedApis,
    mutationFn: (apis, idempotencyKey: string) =>
      apis.billingApi.recharge({
        amount: DEMO_RECHARGE_AMOUNT,
        idempotencyKey,
      }),
    onSuccess: async () => {
      toast.success('Recharge submitted')
      await walletQuery.refresh()
    },
    onError: () => {
      toast.error('Recharge failed')
    },
  })

  const handleRecharge = () => {
    rechargeMutation.mutate(crypto.randomUUID())
  }

  return {
    wallet: walletQuery.data,
    loading: walletQuery.loading,
    canWrite,
    rechargePending: rechargeMutation.isPending,
    handleRecharge,
  }
}
