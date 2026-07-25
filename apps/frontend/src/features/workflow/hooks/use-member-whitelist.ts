import { useCallback } from 'react'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'

export function useMemberWhitelist() {
  const apis = useInjectedApis()
  const { memberId } = useSession()

  const resolveAllowedModelIds = useCallback(async (): Promise<string[] | undefined> => {
    const res = await apis.memberApi.list({ page: 1, pageSize: 500 })
    const member = res.items.find((m) => m.id === memberId)
    if (!member) return undefined
    const resolved = await apis.routingApi.resolveWhitelist(member.departmentId)
    return resolved.allowedModelIds
  }, [apis, memberId])

  return { resolveAllowedModelIds }
}
