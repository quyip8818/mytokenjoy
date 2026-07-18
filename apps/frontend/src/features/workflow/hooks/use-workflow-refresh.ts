import { useCallback } from 'react'
import { useQueryClient, type QueryKey } from '@tanstack/react-query'
import type { WorkflowId, WorkflowPayloadMap } from '../types'
import { useWorkflow } from './use-workflow'

type RefreshHandler = () => void | Promise<void>

export interface WorkflowRefreshOptions {
  refresh?: RefreshHandler
  invalidateKeys?: QueryKey[]
  flashRow?: (id: string) => void
}

export function useWorkflowRefresh(
  refreshOrOptions: RefreshHandler | WorkflowRefreshOptions,
  flashRow?: (id: string) => void,
) {
  const { open } = useWorkflow()
  const queryClient = useQueryClient()

  const resolvedOptions: WorkflowRefreshOptions =
    typeof refreshOrOptions === 'function'
      ? { refresh: refreshOrOptions, flashRow }
      : refreshOrOptions

  const { refresh, invalidateKeys, flashRow: highlightRow } = resolvedOptions

  const runRefresh = useCallback(async () => {
    if (invalidateKeys?.length) {
      await Promise.all(
        invalidateKeys.map((queryKey) => queryClient.invalidateQueries({ queryKey })),
      )
    }
    if (refresh) {
      await refresh()
    }
  }, [invalidateKeys, refresh, queryClient])

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
            void runRefresh()
            if (resultId != null && highlightRow) {
              highlightRow(resultId)
            }
          },
        } as WorkflowPayloadMap[T],
        title,
      )
    },
    [open, highlightRow, runRefresh],
  )

  return { openWithRefresh, open, refresh: runRefresh }
}
