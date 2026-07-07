import { ApiError } from '@/api/client'

export function workflowErrorMessage(err: unknown, fallback: string): string {
  return err instanceof ApiError ? err.message : fallback
}
