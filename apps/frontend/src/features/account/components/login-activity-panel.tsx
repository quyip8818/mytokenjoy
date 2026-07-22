import { Button } from '@/components/ui/button'
import type { LoginActivityResponse } from '@/api/account'

const PAGE_SIZE = 20

interface LoginActivityPanelProps {
  data: LoginActivityResponse | null
  loading: boolean
  offset: number
  onOffsetChange: (offset: number) => void
}

export function LoginActivityPanel({ data, loading, offset, onOffsetChange }: LoginActivityPanelProps) {
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
              <p className="text-xs text-muted-foreground">
                IP: {item.ip}
              </p>
            </div>
            <span className="text-xs text-muted-foreground whitespace-nowrap">
              {formatTime(item.time)}
            </span>
          </div>
        ))}
      </div>

      {(hasPrev || hasNext) && (
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">
            共 {data.total} 条
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={!hasPrev}
              onClick={() => onOffsetChange(Math.max(0, offset - PAGE_SIZE))}
            >
              上一页
            </Button>
            <Button
              variant="outline"
              size="sm"
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

function formatTime(iso: string): string {
  const d = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHours = Math.floor(diffMin / 60)
  if (diffHours < 24) return `${diffHours} 小时前`
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return `${diffDays} 天前`
  return d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
