import { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { Archive, Bell } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useNotifications, useUnreadCount, CATEGORY_MAP, formatTimeAgo, getActionUrl } from '@/features/notifications'
import { useApis } from '@/api/use-apis'
import { useQueryClient } from '@tanstack/react-query'
import type { NotificationItem } from '@/api/types'

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
      className={`group relative flex w-full items-start gap-2.5 border-b border-border px-3 py-2.5 text-left transition-colors hover:bg-muted/60 ${
        isUnread ? 'border-l-2 border-l-blue-500 bg-blue-50/40 dark:bg-blue-950/15' : ''
      }`}
      role="button"
      tabIndex={0}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter') handleClick()
      }}
    >
      {/* Category icon */}
      {Icon && <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${color}`} />}

      {/* Content */}
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-1.5">
          <span
            className={`truncate text-sm ${isUnread ? 'font-medium text-foreground' : 'font-normal text-muted-foreground'}`}
          >
            {notification.title}
          </span>
          <div className="flex shrink-0 items-center gap-1">
            {notification.groupCount && notification.groupCount > 1 && (
              <span className="rounded bg-muted px-1 py-0.5 text-[10px] text-muted-foreground">
                ×{notification.groupCount}
              </span>
            )}
            <span className="text-[10px] text-muted-foreground">
              {formatTimeAgo(notification.createdAt)}
            </span>
          </div>
        </div>
        {notification.body && (
          <span className="line-clamp-1 text-xs text-muted-foreground">{notification.body}</span>
        )}
      </div>

      {/* Archive button on hover */}
      <Button
        variant="ghost"
        size="icon"
        className="absolute right-2 top-2.5 h-5 w-5 opacity-0 transition-opacity group-hover:opacity-100"
        onClick={(e) => {
          e.stopPropagation()
          onArchive(notification.id)
        }}
        title="归档"
      >
        <Archive className="h-3 w-3" />
      </Button>
    </div>
  )
}

export function NotificationInbox() {
  const navigate = useNavigate()
  const { data: notifications } = useNotifications()
  const { data: unreadData } = useUnreadCount()
  const { notificationApi } = useApis()
  const queryClient = useQueryClient()
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

  const handleMarkAllRead = useCallback(async () => {
    await notificationApi.markAllRead()
    invalidate()
  }, [notificationApi, invalidate])

  const handleArchive = useCallback(
    async (id: string) => {
      await notificationApi.archive(id)
      invalidate()
    },
    [notificationApi, invalidate],
  )

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative h-8 w-8">
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-medium text-white">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
          <span className="sr-only">通知</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[360px] p-0" align="end">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-3 py-2.5">
          <span className="text-sm font-semibold">通知</span>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-auto px-2 py-1 text-xs"
              onClick={handleMarkAllRead}
            >
              全部已读
            </Button>
          )}
        </div>

        {/* List */}
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

        {/* Footer */}
        <div className="border-t border-border">
          <button
            type="button"
            className="flex w-full items-center justify-center py-2.5 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            onClick={() => navigate('/notifications')}
          >
            查看全部通知 →
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
