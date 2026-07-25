import { useInjectedQuery, queryKeys } from '@/features/query'

export function useSupplierOptions() {
  const { data } = useInjectedQuery({
    queryKey: queryKeys.suppliers.options(),
    queryFn: (apis) => apis.suppliersApi.options(),
  })
  return data ?? []
}
