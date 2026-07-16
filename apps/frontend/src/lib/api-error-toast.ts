import { toast } from 'sonner'
import { ApiError } from '@/api/client'

/**
 * Extracts a user-facing error message from an unknown error.
 * Returns the API error message if available, otherwise the fallback.
 */
export function apiErrorMessage(err: unknown, fallback: string): string {
  return err instanceof ApiError ? err.message : fallback
}

/**
 * Executes an async action; on failure shows a toast and re-throws.
 * Use in event handlers and mutations to avoid repetitive try/catch boilerplate.
 *
 * @example
 * const save = () => withErrorToast(
 *   () => api.updateItem(id, data),
 *   '保存失败'
 * )
 */
export async function withErrorToast<T>(fn: () => Promise<T>, fallbackMessage: string): Promise<T> {
  try {
    return await fn()
  } catch (err) {
    toast.error(apiErrorMessage(err, fallbackMessage))
    throw err
  }
}
