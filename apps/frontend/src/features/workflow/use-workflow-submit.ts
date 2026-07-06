import { useCallback } from 'react'
import { toast } from 'sonner'
import type { QueryKey } from '@tanstack/react-query'
import { useInjectedMutation } from '@/features/query/use-injected-mutation'
import { useWorkflow } from './use-workflow'

interface UseWorkflowSubmitOptions {
  onSubmit: () => Promise<void>
  onSuccess?: () => void
  successMessage?: string
  errorMessage?: string
  validate?: () => string | null
  invalidateKeys?: QueryKey[]
}

export function useWorkflowSubmit({
  onSubmit,
  onSuccess,
  successMessage,
  errorMessage = '操作失败',
  validate,
  invalidateKeys,
}: UseWorkflowSubmitOptions) {
  const { closeAll } = useWorkflow()
  const { mutateAsync, isPending } = useInjectedMutation<void, void>({
    mutationFn: async () => {
      await onSubmit()
    },
    invalidateKeys,
  })

  const submit = useCallback(async () => {
    const validationError = validate?.()
    if (validationError) {
      return { ok: false as const, error: validationError }
    }
    try {
      await mutateAsync()
      if (successMessage) toast.success(successMessage)
      onSuccess?.()
      closeAll()
      return { ok: true as const }
    } catch {
      toast.error(errorMessage)
      return { ok: false as const, error: errorMessage }
    }
  }, [closeAll, errorMessage, mutateAsync, onSuccess, successMessage, validate])

  return { submit, submitting: isPending }
}
