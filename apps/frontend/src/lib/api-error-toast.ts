import { toast } from '@/lib/toast'
import { ApiError } from '@/api/client'

/**
 * Returns the ApiError directly (preserves code for deep link matching)
 * or a fallback string for non-API errors.
 */
export function apiErrorMessage(err: unknown, fallback: string): ApiError | string {
  return err instanceof ApiError ? err : fallback
}

/**
 * Executes an async action; on failure shows a toast and re-throws.
 * Use in event handlers and mutations to avoid repetitive try/catch boilerplate.
 */
export async function withErrorToast<T>(fn: () => Promise<T>, fallbackMessage: string): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    toast.error(apiErrorMessage(err, fallbackMessage))
    throw err
  }
}
