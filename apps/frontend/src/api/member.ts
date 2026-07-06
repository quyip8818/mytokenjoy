import { request } from './client'
import type { MemberDashboardView } from './types/member'

export const meApi = {
  getDashboard: () => request<MemberDashboardView>('/me/dashboard'),
}
