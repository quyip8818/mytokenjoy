import { useCallback } from 'react'
import type { WorkflowId, WorkflowPayloadMap } from '@/features/workflow/types'
import { useWorkflow } from '@/features/workflow/use-workflow'

export function useWorkflowRefresh(
  refresh: () => void | Promise<void>,
  flashRow?: (id: string) => void,
) {
  const { open } = useWorkflow()

  const openWithRefresh = useCallback(
    <T extends WorkflowId>(id: T, payload?: WorkflowPayloadMap[T], title?: string) => {
      const { onSuccess, ...rest } = (payload ?? {}) as WorkflowPayloadMap[T] & {
        onSuccess?: (id?: string) => void
      }
      open(
        id,
        {
          ...rest,
          onSuccess: (resultId?: string) => {
            onSuccess?.(resultId)
            void refresh()
            if (resultId && flashRow) flashRow(resultId)
          },
        } as WorkflowPayloadMap[T],
        title,
      )
    },
    [open, refresh, flashRow],
  )

  return { openWithRefresh, open }
}
