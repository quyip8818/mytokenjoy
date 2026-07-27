export interface NotificationItem {
  id: string
  eventType: string
  channel: string
  category: string
  title: string
  body: string
  payload: Record<string, unknown>
  groupKey?: string
  groupCount?: number
  status: string // active / archived / deleted
  createdAt: string
  readAt: string | null
  updatedAt: string
}

export interface NotificationListResponse {
  items: NotificationItem[]
  nextCursor: string | null
  hasMore: boolean
}

export interface NotificationUnreadCount {
  count: number
}

export interface NotificationCapabilities {
  channels: string[]
  emailConfigured: boolean
  smsConfigured: boolean
  inAppConfigured: boolean
}

export interface NotificationPreferenceEntry {
  category: string
  channel: string
  enabled: boolean
}

export interface NotificationPreferencesResponse {
  preferences: NotificationPreferenceEntry[]
}

export interface UpdatePreferencesRequest {
  preferences: NotificationPreferenceEntry[]
}

export interface NotificationListParams {
  limit?: number
  cursor?: string
  category?: string
  status?: string // 'unread' | 'read' | ''
  archived?: boolean
  grouped?: boolean
  group_key?: string
}
