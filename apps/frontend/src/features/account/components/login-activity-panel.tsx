import { Button } from '@/components/ui/button'
import { formatTimeAgo } from '@/lib/date'
import type { LoginActivityResponse } from '@/api/types'

const PAGE_SIZE = 20

interface LoginActivityPanelProps {
  data: LoginActivityResponse | null
  loading: boolean
  offset: number
  onOffsetChange: (offset: number) => void
}

export function LoginActivityPanel({
  data,
  loading,
  offset,
  onOffsetChange,
}: LoginActivityPanelProps) {
  if (loading && !data) {
    return (
      <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
        加载中…
      </div>
    )
  }

  if (!data || data.items.length === 0) {
    return (
      <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
        暂无登录记录
      </div>
    )
  }

  const hasNext = offset + PAGE_SIZE < data.total
  const hasPrev = offset > 0

  return (
    <div className="space-y-3">
      <div className="rounded-lg border border-border bg-card divide-y divide-border">
        {data.items.map((item, i) => (
          <div key={`${item.time}-${i}`} className="flex items-center justify-between px-4 py-3">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <span className="text-sm">{formatUserAgent(item.userAgent)}</span>
                {item.current && (
                  <span className="rounded bg-green-100 px-1.5 py-0.5 text-[10px] font-medium text-green-700">
                    当前
                  </span>
                )}
              </div>
              <p className="text-xs text-muted-foreground">IP: {item.ip}</p>
            </div>
            <span className="text-xs text-muted-foreground whitespace-nowrap">
              {formatTimeAgo(item.time)}
            </span>
          </div>
        ))}
      </div>

      {(hasPrev || hasNext) && (
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">共 {data.total} 条</span>
          <div className="flex gap-2">
            <Button
              variant="outline"

              disabled={!hasPrev}
              onClick={() => onOffsetChange(Math.max(0, offset - PAGE_SIZE))}
            >
              上一页
            </Button>
            <Button
              variant="outline"

              disabled={!hasNext}
              onClick={() => onOffsetChange(offset + PAGE_SIZE)}
            >
              下一页
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function formatUserAgent(ua: string): string {
  // Basic extraction: browser + OS
  const browser = extractBrowser(ua)
  const os = extractOS(ua)
  if (browser && os) return `${browser} / ${os}`
  if (browser) return browser
  if (os) return os
  return ua.slice(0, 40) + (ua.length > 40 ? '…' : '')
}

function extractBrowser(ua: string): string | null {
  if (ua.includes('Edg/')) return 'Edge'
  if (ua.includes('Chrome/')) return 'Chrome'
  if (ua.includes('Firefox/')) return 'Firefox'
  if (ua.includes('Safari/') && !ua.includes('Chrome')) return 'Safari'
  return null
}

function extractOS(ua: string): string | null {
  if (ua.includes('Windows')) return 'Windows'
  if (ua.includes('Mac OS X') || ua.includes('Macintosh')) return 'macOS'
  if (ua.includes('Linux') && !ua.includes('Android')) return 'Linux'
  if (ua.includes('Android')) return 'Android'
  if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS'
  return null
}
