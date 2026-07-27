import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { Archive, ArrowUpRight, Bell } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import {
  useNotifications,
  useUnreadCount,
  CATEGORY_MAP,
  formatTimeAgo,
  getActionUrl,
} from '@/features/notifications'
import { useApis } from '@/api/use-apis'
import { useQueryClient } from '@tanstack/react-query'
import type { NotificationItem } from '@/api/types'

// --- Internal list item (Popover only) ---

function NotificationItemRow({
  notification,
  onRead,
  onArchive,
}: {
  notification: NotificationItem
  onRead: (id: string) => void
  onArchive: (id: string) => void
}) {
  const navigate = useNavigate()
  const isUnread = !notification.readAt
  const config = CATEGORY_MAP[notification.category]
  const Icon = config?.icon
  const color = config?.color ?? 'text-muted-foreground'

  const handleClick = () => {
    if (isUnread) onRead(notification.id)
    const url = getActionUrl(notification)
    if (url) navigate(url)
  }

  return (
    <div
      className={cn(
        'group flex w-full items-start gap-3 border-b border-border px-4 py-3 text-left transition-colors hover:bg-muted/60',
        isUnread && 'border-l-[3px] border-l-primary bg-accent/40',
      )}
      role="button"
      tabIndex={0}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter') handleClick()
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
            <span className="text-[10px] text-muted-foreground group-hover:hidden">
              {formatTimeAgo(notification.createdAt)}
            </span>
            <Button
              variant="ghost"
              size="icon"
              className="hidden h-5 w-5 group-hover:inline-flex"
              onClick={(e) => {
                e.stopPropagation()
                onArchive(notification.id)
              }}
              title="归档"
            >
              <Archive className="h-3 w-3" />
            </Button>
          </div>
        </div>
        {notification.body && (
          <p className="mt-0.5 line-clamp-1 text-xs text-muted-foreground">{notification.body}</p>
        )}
      </div>
    </div>
  )
}

// --- Bell trigger + Popover ---

export function NotificationInbox() {
  const navigate = useNavigate()
  const { data: notifications } = useNotifications()
  const { data: unreadData } = useUnreadCount()
  const { notificationApi } = useApis()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const unreadCount = unreadData?.count ?? 0

  const invalidate = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['notifications'] })
  }, [queryClient])

  const handleMarkRead = useCallback(
    async (id: string) => {
      await notificationApi.markRead(id)
      invalidate()
    },
    [notificationApi, invalidate],
  )

  const handleArchive = useCallback(
    async (id: string) => {
      await notificationApi.archive(id)
      invalidate()
    },
    [notificationApi, invalidate],
  )

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <TooltipProvider delayDuration={300}>
        {/* ponytail: disable tooltip when popover is open to prevent Radix overlap bug */}
        <Tooltip open={open ? false : undefined}>
          <TooltipTrigger asChild>
            <PopoverTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="relative h-9 w-9 rounded-md border border-border hover:bg-muted"
              >
                <Bell
                  className={cn(
                    'h-[18px] w-[18px] transition-colors',
                    unreadCount > 0 ? 'text-foreground' : 'text-foreground/60',
                  )}
                />
                {unreadCount > 0 && (
                  <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-medium tabular-nums text-primary-foreground">
                    {unreadCount > 99 ? '99+' : unreadCount}
                  </span>
                )}
                <span className="sr-only">通知</span>
              </Button>
            </PopoverTrigger>
          </TooltipTrigger>
          <TooltipContent side="bottom" className="text-xs">
            {unreadCount > 0 ? `${unreadCount} 条未读通知` : '通知'}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <PopoverContent className="w-[360px] p-0" align="end">
        <div className="relative flex items-center justify-center border-b border-border px-3 py-2.5">
          <span className="text-sm font-semibold">通知</span>
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 h-6 w-6 text-muted-foreground hover:text-foreground"
            onClick={() => {
              setOpen(false)
              navigate('/notifications')
            }}
            title="查看全部通知"
          >
            <ArrowUpRight className="h-3.5 w-3.5" />
          </Button>
        </div>

        <ScrollArea className="max-h-[400px]">
          {!notifications || notifications.length === 0 ? (
            <div className="flex h-40 flex-col items-center justify-center gap-1.5">
              <Bell className="h-8 w-8 text-muted-foreground/30" strokeWidth={1} />
              <span className="text-sm text-muted-foreground">暂无新通知</span>
            </div>
          ) : (
            notifications.map((n) => (
              <NotificationItemRow
                key={n.id}
                notification={n}
                onRead={handleMarkRead}
                onArchive={handleArchive}
              />
            ))
          )}
        </ScrollArea>

        <div className="border-t border-border">
          <button
            type="button"
            className="flex w-full items-center justify-center py-2.5 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            onClick={() => {
              setOpen(false)
              navigate('/notifications')
            }}
          >
            查看全部通知 →
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
