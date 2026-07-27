// === 页面入口（route page 消费）===
export { useNotificationsPage } from './hooks/use-notifications-page'
export { NotificationsPageShell } from './components/notifications-page-shell'

// === 通知中心页面 ===
export { NotificationCenter } from './components/notification-center'
export { useNotificationInbox } from './hooks/use-notification-inbox'

// === 跨 feature/layout 共享 ===
// consumed by: components/layout/notification-inbox
export { useNotifications, useUnreadCount } from './hooks/use-notifications'
// consumed by: components/layout/app-providers
export { NotificationProvider } from './notification-provider'

// === 工具函数 ===
export { getActionUrl } from './lib/get-action-url'
export { CATEGORY_MAP, ALL_CATEGORIES } from './lib/category-config'
export { formatTimeAgo } from './lib/format-time'
