// Notification module shared contracts.
// These types define the common vocabulary between frontend and backend.

// --- Channels ---

export const NOTIFICATION_CHANNEL = {
  EMAIL: 'email',
  SMS: 'sms',
  IN_APP: 'in_app',
  LOG: 'log',
  WEBHOOK: 'webhook',
} as const

export type NotificationChannel =
  (typeof NOTIFICATION_CHANNEL)[keyof typeof NOTIFICATION_CHANNEL]

// --- Priority ---

export const NOTIFICATION_PRIORITY = {
  CRITICAL: 'critical',
  HIGH: 'high',
  NORMAL: 'normal',
  LOW: 'low',
} as const

export type NotificationPriority =
  (typeof NOTIFICATION_PRIORITY)[keyof typeof NOTIFICATION_PRIORITY]

// --- Categories ---

export const NOTIFICATION_CATEGORY = {
  BUDGET_ALERT: 'budget_alert',
  KEY_EXPIRATION: 'key_expiration',
  USAGE_REPORT: 'usage_report',
  SECURITY_EVENT: 'security_event',
  SYSTEM_MAINTENANCE: 'system_maintenance',
  OVERRUN: 'overrun',
} as const

export type NotificationCategory =
  (typeof NOTIFICATION_CATEGORY)[keyof typeof NOTIFICATION_CATEGORY]

// --- Event Types ---

export const NOTIFICATION_EVENT = {
  SYNC_THRESHOLD_EXCEEDED: 'sync_threshold_exceeded',
  OVERRUN_BLOCKED: 'overrun_blocked',
  OVERDRAFT_EXPANDED: 'overdraft_expanded',
  BUDGET_ALERT_REACHED: 'budget_alert_reached',
  KEY_EXPIRED: 'key_expired',
  KEY_EXPIRING_SOON: 'key_expiring_soon',
  USAGE_WEEKLY_REPORT: 'usage_weekly_report',
  SECURITY_LOGIN_NEW_DEVICE: 'security_login_new_device',
  SYSTEM_MAINTENANCE_SCHEDULED: 'system_maintenance_scheduled',
} as const

export type NotificationEventType =
  (typeof NOTIFICATION_EVENT)[keyof typeof NOTIFICATION_EVENT]

// --- Status ---

export const NOTIFICATION_STATUS = {
  PENDING: 'pending',
  SENT: 'sent',
  FAILED: 'failed',
  READ: 'read',
} as const

export type NotificationStatus =
  (typeof NOTIFICATION_STATUS)[keyof typeof NOTIFICATION_STATUS]

// --- Notification Event (trigger payload) ---

export interface NotificationEvent {
  eventType: NotificationEventType | string
  recipientId: string
  companyId: number
  payload: Record<string, unknown>
  metadata?: NotificationMetadata
}

export interface NotificationMetadata {
  deduplicationKey?: string
  priority?: NotificationPriority
  groupKey?: string
  category?: NotificationCategory
}

// --- User Preference ---

export interface NotificationPreferenceEntry {
  category: NotificationCategory | string
  channel: NotificationChannel
  enabled: boolean
}

export interface NotificationPreferences {
  userId: string
  preferences: NotificationPreferenceEntry[]
  globalMute: boolean
  quietHours?: QuietHours
}

export interface QuietHours {
  start: string // HH:mm
  end: string // HH:mm
  timezone: string
}

// --- Notification Item (for inbox display) ---

export interface NotificationItem {
  id: string
  eventType: string
  channel: string
  title: string
  body: string
  status: NotificationStatus
  createdAt: string // ISO 8601
  readAt?: string | null
  payload?: Record<string, unknown>
}

// --- Capabilities response ---

export interface NotificationCapabilities {
  channels: NotificationChannel[]
  emailConfigured: boolean
  smsConfigured: boolean
  inAppConfigured: boolean
}

// --- Priority fallback chains ---

export const PRIORITY_FALLBACK_CHAIN: Record<NotificationPriority, NotificationChannel[]> = {
  critical: ['sms', 'email', 'in_app'],
  high: ['email', 'in_app'],
  normal: ['in_app'],
  low: ['in_app'],
}

// --- Category to default channels mapping ---

export const CATEGORY_DEFAULT_CHANNELS: Record<NotificationCategory, NotificationChannel[]> = {
  budget_alert: ['email', 'in_app'],
  key_expiration: ['email', 'in_app'],
  usage_report: ['email', 'in_app'],
  security_event: ['email', 'sms', 'in_app'],
  system_maintenance: ['in_app'],
  overrun: ['email', 'in_app'],
}
