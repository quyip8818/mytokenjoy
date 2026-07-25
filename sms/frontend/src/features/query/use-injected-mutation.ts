import { useMutation, useQueryClient, type QueryKey } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'

export function useInjectedMutation<TData = void, TVariables = void>({
  injectedApis,
  mutationFn,
  invalidateKeys,
  onSuccess,
  onError,
}: {
  injectedApis?: AppApis
  mutationFn: (apis: AppApis, variables: TVariables) => Promise<TData>
  invalidateKeys?: QueryKey[]
  onSuccess?: (data: TData) => void
  onError?: (error: Error) => void
}) {
  const apis = useInjectedApis(injectedApis)
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (variables: TVariables) => mutationFn(apis, variables),
    onSuccess: async (data) => {
      if (invalidateKeys) {
        await Promise.all(invalidateKeys.map((k) => qc.invalidateQueries({ queryKey: k })))
      }
      onSuccess?.(data)
    },
    onError: (error) => onError?.(error as Error),
  })
}
