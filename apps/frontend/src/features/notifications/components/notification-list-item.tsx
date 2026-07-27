import { Archive, ArchiveRestore, Check, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
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
      className={cn(
        'group flex cursor-pointer items-start gap-3 px-4 py-3.5 transition-colors hover:bg-muted/50',
        isUnread && 'border-l-[3px] border-l-primary bg-accent/40',
      )}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onClick()
      }}
    >
      {Icon && <Icon className={cn('mt-0.5 h-4 w-4 shrink-0', color)} />}

      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <span
            className={cn(
              'truncate text-sm',
              isUnread ? 'font-medium text-foreground' : 'font-normal text-muted-foreground',
            )}
          >
            {notification.title}
          </span>
          <div className="flex shrink-0 items-center gap-1.5">
            {notification.groupCount && notification.groupCount > 1 && (
              <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-muted px-1.5 text-[10px] font-medium tabular-nums text-muted-foreground">
                {notification.groupCount}
              </span>
            )}
            {/* Time: default visible, hidden on hover */}
            <span className="text-xs text-muted-foreground group-hover:hidden">
              {formatTimeAgo(notification.createdAt)}
            </span>
            {/* Actions: visible on hover, replace time */}
            <TooltipProvider delayDuration={200}>
              <div className="hidden items-center gap-1 group-hover:flex">
                {tab === 'inbox' && isUnread && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="outline"
                        size="icon-sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          onMarkRead()
                        }}
                      >
                        <Check className="h-3.5 w-3.5" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent side="bottom" className="text-xs">标记已读</TooltipContent>
                  </Tooltip>
                )}
                {tab === 'inbox' ? (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="outline"
                        size="icon-sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          onArchive()
                        }}
                      >
                        <Archive className="h-3.5 w-3.5" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent side="bottom" className="text-xs">归档</TooltipContent>
                  </Tooltip>
                ) : (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="outline"
                        size="icon-sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          onUnarchive()
                        }}
                      >
                        <ArchiveRestore className="h-3.5 w-3.5" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent side="bottom" className="text-xs">恢复</TooltipContent>
                  </Tooltip>
                )}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="destructive"
                      size="icon-sm"
                      onClick={(e) => {
                        e.stopPropagation()
                        onDelete()
                      }}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="text-xs">删除</TooltipContent>
                </Tooltip>
              </div>
            </TooltipProvider>
          </div>
        </div>
        {notification.body && (
          <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">{notification.body}</p>
        )}
      </div>
    </div>
  )
}
