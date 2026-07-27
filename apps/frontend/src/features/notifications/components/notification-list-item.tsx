import { Archive, ArchiveRestore, Check, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { CATEGORY_MAP } from '../lib/category-config'
import { formatTimeAgo } from '../lib/format-time'
import type { NotificationItem } from '@/api/types'
import type { InboxTab } from '../hooks/use-notification-inbox'

interface Props {
  notification: NotificationItem
  tab: InboxTab
  onClick: () => void
  onArchive: () => void
  onUnarchive: () => void
  onDelete: () => void
  onMarkRead: () => void
}

export function NotificationListItem({
  notification,
  tab,
  onClick,
  onArchive,
  onUnarchive,
  onDelete,
  onMarkRead,
}: Props) {
  const isUnread = !notification.readAt
  const config = CATEGORY_MAP[notification.category]
  const Icon = config?.icon
  const color = config?.color ?? 'text-muted-foreground'

  return (
    <div
      className={`group relative flex cursor-pointer items-start gap-3 px-4 py-3 transition-colors hover:bg-muted/60 ${
        isUnread ? 'border-l-2 border-l-blue-500 bg-blue-50/40 dark:bg-blue-950/15' : ''
      }`}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onClick()
      }}
    >
      {/* Category icon */}
      {Icon && <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${color}`} />}

      {/* Content */}
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <span
            className={`truncate text-sm ${isUnread ? 'font-medium text-foreground' : 'font-normal text-muted-foreground'}`}
          >
            {notification.title}
          </span>
          <div className="flex shrink-0 items-center gap-1.5">
            {notification.groupCount && notification.groupCount > 1 && (
              <span className="rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
                {notification.groupCount}
              </span>
            )}
            <span className="text-xs text-muted-foreground">
              {formatTimeAgo(notification.createdAt)}
            </span>
          </div>
        </div>
        {notification.body && (
          <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">{notification.body}</p>
        )}
      </div>

      {/* Action buttons (visible on hover, always visible on touch) */}
      <div className="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100 md:opacity-0">
        {tab === 'inbox' && isUnread && (
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={(e) => {
              e.stopPropagation()
              onMarkRead()
            }}
            title="标记已读"
          >
            <Check className="h-3.5 w-3.5" />
          </Button>
        )}
        {tab === 'inbox' ? (
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={(e) => {
              e.stopPropagation()
              onArchive()
            }}
            title="归档"
          >
            <Archive className="h-3.5 w-3.5" />
          </Button>
        ) : (
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={(e) => {
              e.stopPropagation()
              onUnarchive()
            }}
            title="恢复"
          >
            <ArchiveRestore className="h-3.5 w-3.5" />
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 text-destructive hover:bg-destructive/10"
          onClick={(e) => {
            e.stopPropagation()
            onDelete()
          }}
          title="删除"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  )
}
