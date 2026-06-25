import { useCallback } from 'react'
import { useApis } from '@/api/use-apis'
import { useDemoRole } from '@/features/demo'
import type { WorkflowComponentProps } from './types'

export function useMemberWhitelist() {
  const apis = useApis()
  const { memberId } = useDemoRole()

  const resolveWhitelist = useCallback(async (): Promise<string[] | undefined> => {
    const res = await apis.memberApi.list({ page: 1, pageSize: 500 })
    const member = res.items.find((m) => m.id === memberId)
    if (!member) return undefined
    const resolved = await apis.routingApi.resolveWhitelist(member.departmentId)
    return resolved.allowedModels
  }, [apis, memberId])

  return { resolveWhitelist }
}

export async function pushModelPicker(
  onPush: WorkflowComponentProps['onPush'],
  resolveWhitelist: () => Promise<string[] | undefined>,
  {
    selectedModels,
    onConfirm,
    onSetDirty,
  }: {
    selectedModels: string[]
    onConfirm: (picked: string[]) => void
    onSetDirty?: (dirty: boolean) => void
  },
) {
  const parentWhitelist = await resolveWhitelist()
  onPush('model-picker', {
    selectedModels,
    parentWhitelist,
    onConfirm: (picked: string[]) => {
      onConfirm(picked)
      onSetDirty?.(true)
    },
  })
}
