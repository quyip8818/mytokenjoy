import { Bell, CheckCircle2, Search } from 'lucide-react'
import type { InboxTab } from '../hooks/use-notification-inbox'

interface Props {
  tab: InboxTab
  hasFilter: boolean
}

export function NotificationEmptyState({ tab, hasFilter }: Props) {
  if (hasFilter) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Search className="mb-3 h-12 w-12 text-muted-foreground/40" strokeWidth={1} />
        <p className="text-sm text-muted-foreground">没有符合条件的通知</p>
        <p className="mt-1 text-xs text-muted-foreground/70">试试调整筛选条件</p>
      </div>
    )
  }

  if (tab === 'archived') {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <CheckCircle2 className="mb-3 h-12 w-12 text-muted-foreground/40" strokeWidth={1} />
        <p className="text-sm text-muted-foreground">没有已归档的通知</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <Bell className="mb-3 h-12 w-12 text-muted-foreground/40" strokeWidth={1} />
      <p className="text-sm text-muted-foreground">暂无通知</p>
      <p className="mt-1 text-xs text-muted-foreground/70">有新动态时会在这里提醒</p>
    </div>
  )
}
