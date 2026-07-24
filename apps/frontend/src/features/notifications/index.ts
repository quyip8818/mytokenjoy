// === 页面入口（route page 消费）===
export { useNotificationsPage } from './hooks/use-notifications-page'
export { NotificationsPageShell } from './components/notifications-page-shell'

// === 跨 feature/layout 共享 ===
// consumed by: components/layout/notification-inbox
export { useNotifications, useUnreadCount } from './hooks/use-notifications'
// consumed by: components/layout/app-providers
export { NotificationProvider } from './notification-provider'
