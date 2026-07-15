export interface NotificationItem {
  id: string
  eventType: string
  channel: string
  title: string
  body: string
  status: string
  createdAt: string
  readAt: string | null
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
