import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useApis } from '@/api/use-apis'

export const loginActivityKeys = {
  list: (offset: number) => ['account', 'loginActivity', offset] as const,
}

export function useLoginActivityPage() {
  const { accountApi } = useApis()
  const [offset, setOffset] = useState(0)

  const query = useQuery({
    queryKey: loginActivityKeys.list(offset),
    queryFn: () => accountApi.getLoginActivity({ limit: 20, offset }),
  })

  return {
    data: query.data ?? null,
    loading: query.isLoading,
    offset,
    setOffset,
  }
}

export type LoginActivityPageState = ReturnType<typeof useLoginActivityPage>
