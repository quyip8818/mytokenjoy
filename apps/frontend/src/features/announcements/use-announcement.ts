import { useCallback, useState } from 'react'
import { useSession } from '@/features/session'
import { announcements } from './announcements'
import type { AnnouncementConfig } from './types'

const STORAGE_PREFIX = 'announcement-dismissed-'

function isDismissed(id: string, companyId: string): boolean {
  return localStorage.getItem(`${STORAGE_PREFIX}${id}-${companyId}`) === '1'
}

function dismiss(id: string, companyId: string): void {
  localStorage.setItem(`${STORAGE_PREFIX}${id}-${companyId}`, '1')
}

/**
 * Returns the first undismissed announcement matching current session, or null.
 * ponytail: only one announcement at a time — simplest UX.
 * 升级路径：支持优先级队列 + 频次控制。
 */
export function useAnnouncement(): {
  current: AnnouncementConfig | null
  open: boolean
  close: () => void
} {
  const session = useSession()
  const [dismissed, setDismissed] = useState<string | null>(null)

  // ponytail: derive current announcement from session state — no effect needed
  const current =
    !session.loading && session.companyId
      ? (announcements.find(
          (a) =>
            a.id !== dismissed && a.shouldShow(session) && !isDismissed(a.id, session.companyId),
        ) ?? null)
      : null

  const open = current !== null

  const close = useCallback(() => {
    if (current) {
      dismiss(current.id, session.companyId)
      setDismissed(current.id)
    }
  }, [current, session.companyId])

  return { current, open, close }
}
