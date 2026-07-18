import { useCallback } from 'react'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
import type { WorkflowComponentProps } from '../types'

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

export async function pushModelPicker(
  onPush: WorkflowComponentProps['onPush'],
  resolveAllowedModelIds: () => Promise<string[] | undefined>,
  {
    selectedModelIds,
    onConfirm,
    onSetDirty,
  }: {
    selectedModelIds: string[]
    onConfirm: (picked: string[]) => void
    onSetDirty?: (dirty: boolean) => void
  },
) {
  const parentAllowedModelIds = await resolveAllowedModelIds()
  onPush('model-picker', {
    selectedModelIds,
    parentAllowedModelIds,
    onConfirm: (picked: string[]) => {
      onConfirm(picked)
      onSetDirty?.(true)
    },
  })
}
