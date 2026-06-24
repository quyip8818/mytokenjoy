import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useWorkflow } from './use-workflow'

interface UseWorkflowSubmitOptions {
  onSubmit: () => Promise<void>
  onSuccess?: () => void
  successMessage?: string
  errorMessage?: string
  validate?: () => string | null
}

export function useWorkflowSubmit({
  onSubmit,
  onSuccess,
  successMessage,
  errorMessage = '操作失败',
  validate,
}: UseWorkflowSubmitOptions) {
  const { closeAll } = useWorkflow()
  const [submitting, setSubmitting] = useState(false)

  const submit = useCallback(async () => {
    const validationError = validate?.()
    if (validationError) {
      return { ok: false as const, error: validationError }
    }
    setSubmitting(true)
    try {
      await onSubmit()
      if (successMessage) toast.success(successMessage)
      onSuccess?.()
      closeAll()
      return { ok: true as const }
    } catch {
      toast.error(errorMessage)
      return { ok: false as const, error: errorMessage }
    } finally {
      setSubmitting(false)
    }
  }, [closeAll, errorMessage, onSubmit, onSuccess, successMessage, validate])

  return { submit, submitting }
}
