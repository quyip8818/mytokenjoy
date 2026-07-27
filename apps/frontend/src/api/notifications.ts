import { buildQuery, request } from './client'
import type {
  NotificationCapabilities,
  NotificationListParams,
  NotificationListResponse,
  NotificationPreferencesResponse,
  NotificationUnreadCount,
  UpdatePreferencesRequest,
} from './types'

export const notificationApi = {
  list: (params?: NotificationListParams) =>
    request<NotificationListResponse>(`/notifications${buildQuery(params ?? {})}`),

  unreadCount: () => request<NotificationUnreadCount>('/notifications/unread-count'),

  markRead: (id: string) => request<void>(`/notifications/${id}/read`, { method: 'PATCH' }),

  archive: (id: string) => request<void>(`/notifications/${id}/archive`, { method: 'POST' }),

  unarchive: (id: string) => request<void>(`/notifications/${id}/unarchive`, { method: 'POST' }),

  softDelete: (id: string) => request<void>(`/notifications/${id}/delete`, { method: 'POST' }),

  undelete: (id: string) => request<void>(`/notifications/${id}/undelete`, { method: 'POST' }),

  getCapabilities: () => request<NotificationCapabilities>('/notifications/capabilities'),

  getPreferences: () => request<NotificationPreferencesResponse>('/notifications/preferences'),

  updatePreferences: (data: UpdatePreferencesRequest) =>
    request<void>('/notifications/preferences', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  resetPreferences: () => request<void>('/notifications/preferences/reset', { method: 'POST' }),
}
