import { request } from './client'
import type { SessionContext } from './types'

export const sessionApi = {
  getCurrent: () => request<SessionContext>('/session'),
}
