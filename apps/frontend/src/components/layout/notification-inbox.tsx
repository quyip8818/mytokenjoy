import { useCallback } from 'react'
import { Bell } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useNotifications, useUnreadCount } from '@/features/notifications'
import { useApis } from '@/api/use-apis'
import { useQueryClient } from '@tanstack/react-query'
import type { NotificationItem } from '@/api/types'

function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60_000)

  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `${diffHours} 小时前`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays} 天前`
  return date.toLocaleDateString()
}

function NotificationItemRow({
  notification,
  onRead,
}: {
  notification: NotificationItem
  onRead: (id: string) => void
}) {
  const isUnread = !notification.readAt

  return (
    <button
      type="button"
      className={`flex w-full flex-col gap-0.5 border-b border-border px-3 py-2.5 text-left transition-colors hover:bg-muted ${
        isUnread ? 'bg-blue-50/50 dark:bg-blue-950/20' : ''
      }`}
      onClick={() => {
        if (isUnread) onRead(notification.id)
      }}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="truncate text-sm font-medium text-foreground">{notification.title}</span>
        {isUnread && <span className="h-2 w-2 shrink-0 rounded-full bg-blue-500" />}
      </div>
      {notification.body && (
        <span className="line-clamp-2 text-xs text-muted-foreground">{notification.body}</span>
      )}
      <span className="mt-0.5 text-[10px] text-muted-foreground">
        {formatTimeAgo(notification.createdAt)}
      </span>
    </button>
  )
}

export function NotificationInbox() {
  const { data: notifications } = useNotifications()
  const { data: unreadData } = useUnreadCount()
  const { notificationApi } = useApis()
  const queryClient = useQueryClient()
  const unreadCount = unreadData?.count ?? 0

  const handleMarkRead = useCallback(
    async (id: string) => {
      await notificationApi.markRead(id)
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
    },
    [notificationApi, queryClient],
  )

  const handleMarkAllRead = useCallback(async () => {
    await notificationApi.markAllRead()
    queryClient.invalidateQueries({ queryKey: ['notifications'] })
    queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
  }, [notificationApi, queryClient])

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
      <PopoverContent className="w-80 p-0" align="end">
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
        <ScrollArea className="h-80">
          {!notifications || notifications.length === 0 ? (
            <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              暂无通知
            </div>
          ) : (
            notifications.map((n) => (
              <NotificationItemRow key={n.id} notification={n} onRead={handleMarkRead} />
            ))
          )}
        </ScrollArea>
      </PopoverContent>
    </Popover>
  )
}
