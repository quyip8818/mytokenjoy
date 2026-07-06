import { useMutation, useQueryClient, type QueryKey } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'

type MutationFn<TData, TVariables> = (apis: AppApis, variables: TVariables) => Promise<TData>

export interface UseInjectedMutationOptions<TData, TVariables> {
  injectedApis?: AppApis
  mutationFn: MutationFn<TData, TVariables>
  invalidateKeys?: QueryKey[] | ((variables: TVariables) => QueryKey[])
  onSuccess?: (data: TData, variables: TVariables) => void
  onError?: (error: Error, variables: TVariables) => void
}

export function useInjectedMutation<TData = void, TVariables = void>({
  injectedApis,
  mutationFn,
  invalidateKeys,
  onSuccess,
  onError,
}: UseInjectedMutationOptions<TData, TVariables>) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (variables: TVariables) => mutationFn(apis, variables),
    onSuccess: async (data, variables) => {
      if (invalidateKeys) {
        const keys =
          typeof invalidateKeys === 'function' ? invalidateKeys(variables) : invalidateKeys
        await Promise.all(keys.map((queryKey) => queryClient.invalidateQueries({ queryKey })))
      }
      onSuccess?.(data, variables)
    },
    onError: (error, variables) => {
      onError?.(error, variables)
    },
  })

  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    reset: mutation.reset,
  }
}
