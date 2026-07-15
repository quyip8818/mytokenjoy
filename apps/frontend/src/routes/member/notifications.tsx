import { NotificationsPageShell, useNotificationsPage } from '@/features/notifications'

export default function MemberNotificationsPage() {
  return <NotificationsPageShell {...useNotificationsPage()} />
}
