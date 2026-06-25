import { request, buildQuery } from './client'
import type { SessionContext } from './types'

export const sessionApi = {
  get: (memberId: string) => request<SessionContext>(`/session${buildQuery({ memberId })}`),
}
