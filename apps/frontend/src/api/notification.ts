import { buildQuery, request } from './client'
import type {
  NotificationCapabilities,
  NotificationItem,
  NotificationPreferencesResponse,
  NotificationUnreadCount,
  UpdatePreferencesRequest,
} from './types'

export const notificationApi = {
  list: (params?: { limit?: number; offset?: number }) =>
    request<NotificationItem[]>(`/notifications${buildQuery(params ?? {})}`),

  unreadCount: () => request<NotificationUnreadCount>('/notifications/unread-count'),

  markRead: (id: string) => request<void>(`/notifications/${id}/read`, { method: 'PATCH' }),

  markAllRead: () => request<void>('/notifications/read-all', { method: 'POST' }),

  getCapabilities: () => request<NotificationCapabilities>('/notifications/capabilities'),

  getPreferences: () => request<NotificationPreferencesResponse>('/notifications/preferences'),

  updatePreferences: (data: UpdatePreferencesRequest) =>
    request<void>('/notifications/preferences', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  resetPreferences: () => request<void>('/notifications/preferences/reset', { method: 'POST' }),
}
