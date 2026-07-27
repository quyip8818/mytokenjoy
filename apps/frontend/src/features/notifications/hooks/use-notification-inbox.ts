import { useCallback, useState } from 'react'
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useApis } from '@/api/use-apis'
import type { NotificationListParams } from '@/api/types'

export type InboxTab = 'inbox' | 'archived'
export type StatusFilter = '' | 'unread' | 'read'

export function useNotificationInbox() {
  const { notificationApi } = useApis()
  const queryClient = useQueryClient()

  const [tab, setTab] = useState<InboxTab>('inbox')
  const [category, setCategory] = useState('')
  const [status, setStatus] = useState<StatusFilter>('')

  const queryKey = ['notifications', 'inbox', tab, category, status]

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = useInfiniteQuery({
    queryKey,
    queryFn: async ({ pageParam }) => {
      const params: NotificationListParams = {
        limit: 20,
        archived: tab === 'archived',
        category: category || undefined,
        status: status || undefined,
        grouped: true,
      }
      if (pageParam) params.cursor = pageParam
      return notificationApi.list(params)
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    initialPageParam: undefined as string | undefined,
  })

  const items = data?.pages.flatMap((p) => p.items) ?? []

  const invalidate = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['notifications'] })
  }, [queryClient])

  const markReadMutation = useMutation({
    mutationFn: (id: string) => notificationApi.markRead(id),
    onSuccess: invalidate,
  })

  const markAllReadMutation = useMutation({
    mutationFn: () => notificationApi.markAllRead(),
    onSuccess: invalidate,
  })

  const archiveMutation = useMutation({
    mutationFn: (id: string) => notificationApi.archive(id),
    onSuccess: invalidate,
  })

  const unarchiveMutation = useMutation({
    mutationFn: (id: string) => notificationApi.unarchive(id),
    onSuccess: invalidate,
  })

  const archiveAllMutation = useMutation({
    mutationFn: () => notificationApi.archiveAll(category || undefined),
    onSuccess: invalidate,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => notificationApi.softDelete(id),
    onSuccess: (_, id) => {
      invalidate()
      toast('已删除 1 条通知', {
        action: {
          label: '撤销',
          onClick: () => {
            notificationApi.undelete(id).then(invalidate)
          },
        },
        duration: 5000,
      })
    },
  })

  return {
    // State
    tab,
    setTab,
    category,
    setCategory,
    status,
    setStatus,
    // Data
    items,
    isLoading,
    hasNextPage: hasNextPage ?? false,
    isFetchingNextPage,
    // Actions
    fetchNextPage,
    markRead: markReadMutation.mutate,
    markAllRead: markAllReadMutation.mutate,
    archive: archiveMutation.mutate,
    unarchive: unarchiveMutation.mutate,
    archiveAll: archiveAllMutation.mutate,
    deleteNotification: deleteMutation.mutate,
  }
}
