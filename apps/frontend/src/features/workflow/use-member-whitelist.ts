import { useCallback } from 'react'
import { memberApi } from '@/api/org'
import { routingApi } from '@/api/models'
import { useDemoRole } from '@/features/demo'
import type { WorkflowComponentProps } from './types'

export function useMemberWhitelist() {
  const { memberId } = useDemoRole()

  const resolveWhitelist = useCallback(async (): Promise<string[] | undefined> => {
    const res = await memberApi.list({ page: 1, pageSize: 500 })
    const member = res.items.find((m) => m.id === memberId)
    if (!member) return undefined
    const resolved = await routingApi.resolveWhitelist(member.departmentId)
    return resolved.allowedModels
  }, [memberId])

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
