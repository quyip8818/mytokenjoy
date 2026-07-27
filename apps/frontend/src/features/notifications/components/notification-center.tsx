import { useCallback, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { Archive, Check, Filter, Inbox } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { useNotificationInbox } from '../hooks/use-notification-inbox'
import { useUnreadCount } from '../hooks/use-notifications'
import { NotificationListItem } from './notification-list-item'
import { NotificationEmptyState } from './notification-empty-state'
import { ALL_CATEGORIES } from '../lib/category-config'
import { getActionUrl } from '../lib/get-action-url'
import type { NotificationItem } from '@/api/types'

export function NotificationCenter() {
  const navigate = useNavigate()
  const {
    tab,
    setTab,
    category,
    setCategory,
    status,
    setStatus,
    items,
    isLoading,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    markRead,
    markAllRead,
    archive,
    unarchive,
    archiveAll,
    deleteNotification,
  } = useNotificationInbox()

  const { data: unreadData } = useUnreadCount()
  const unreadCount = unreadData?.count ?? 0

  // Infinite scroll sentinel
  const sentinelRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage()
        }
      },
      { threshold: 0.1 },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const handleClick = useCallback(
    (notification: NotificationItem) => {
      if (!notification.readAt) markRead(notification.id)
      const url = getActionUrl(notification)
      if (url) navigate(url)
    },
    [markRead, navigate],
  )

  return (
    <PageShell>
      <PageHeader
        title="通知中心"
        actions={
          <Button
            variant="ghost"
            size="sm"
            className="text-xs text-muted-foreground"
            onClick={() => navigate('/me/settings')}
          >
            通知偏好设置
          </Button>
        }
      />

      <div className="mx-auto max-w-3xl">
        {/* Tabs */}
        <Tabs value={tab} onValueChange={(v) => setTab(v as 'inbox' | 'archived')}>
          <TabsList className="mb-4">
            <TabsTrigger value="inbox" className="gap-1.5">
              <Inbox className="h-3.5 w-3.5" />
              收件箱
              {unreadCount > 0 && (
                <span className="ml-1 text-xs text-blue-600">●{unreadCount}</span>
              )}
            </TabsTrigger>
            <TabsTrigger value="archived" className="gap-1.5">
              <Archive className="h-3.5 w-3.5" />
              已归档
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {/* Toolbar */}
        <div className="mb-3 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="h-8 w-[130px] text-xs">
                <Filter className="mr-1 h-3 w-3" />
                <SelectValue placeholder="全部类别" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">全部类别</SelectItem>
                {ALL_CATEGORIES.map((c) => (
                  <SelectItem key={c.key} value={c.key}>
                    {c.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {tab === 'inbox' && (
              <Select value={status} onValueChange={(v) => setStatus(v as '' | 'unread' | 'read')}>
                <SelectTrigger className="h-8 w-[100px] text-xs">
                  <SelectValue placeholder="全部状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">全部</SelectItem>
                  <SelectItem value="unread">未读</SelectItem>
                  <SelectItem value="read">已读</SelectItem>
                </SelectContent>
              </Select>
            )}
          </div>

          {tab === 'inbox' && (
            <div className="flex items-center gap-1.5">
              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={() => markAllRead()}
                >
                  <Check className="mr-1 h-3 w-3" />
                  全部已读
                </Button>
              )}
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs"
                onClick={() => archiveAll()}
              >
                <Archive className="mr-1 h-3 w-3" />
                全部归档
              </Button>
            </div>
          )}
        </div>

        {/* List */}
        {isLoading ? (
          <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : items.length === 0 ? (
          <NotificationEmptyState tab={tab} hasFilter={!!category || !!status} />
        ) : (
          <div className="divide-y divide-border rounded-lg border border-border">
            {items.map((item) => (
              <NotificationListItem
                key={item.id}
                notification={item}
                tab={tab}
                onClick={() => handleClick(item)}
                onArchive={() => archive(item.id)}
                onUnarchive={() => unarchive(item.id)}
                onDelete={() => deleteNotification(item.id)}
                onMarkRead={() => markRead(item.id)}
              />
            ))}
          </div>
        )}

        {/* Infinite scroll sentinel */}
        <div ref={sentinelRef} className="h-8" />
        {isFetchingNextPage && (
          <div className="py-4 text-center text-xs text-muted-foreground">加载更多...</div>
        )}
      </div>
    </PageShell>
  )
}
