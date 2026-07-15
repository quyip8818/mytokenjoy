import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { PageShell } from '@/components/layout/page-shell'
import type { useNotificationsPage } from '@/features/notifications'

type NotificationsPageShellProps = ReturnType<typeof useNotificationsPage>

export function NotificationsPageShell({
  categories,
  channels,
  loading,
  saving,
  isChannelEnabled,
  isChannelConfigured,
  togglePreference,
  resetPreferences,
}: NotificationsPageShellProps) {
  if (loading) {
    return (
      <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
        加载中...
      </div>
    )
  }

  return (
    <PageShell
      description={
        <div>
          <h2 className="text-lg font-semibold">通知偏好</h2>
          <p className="text-sm text-muted-foreground">选择每种通知类型的接收渠道</p>
        </div>
      }
      actions={
        <Button variant="outline" size="sm" onClick={resetPreferences} disabled={saving}>
          恢复默认
        </Button>
      }
    >
      <div className="rounded-lg border border-border">
        {/* Header */}
        <div className="grid grid-cols-[1fr_80px_80px_80px] items-center gap-2 border-b border-border bg-muted/50 px-4 py-3">
          <span className="text-xs font-medium text-muted-foreground">类别</span>
          {channels.map((ch) => (
            <span key={ch.key} className="text-center text-xs font-medium text-muted-foreground">
              {ch.label}
              {!isChannelConfigured(ch.key) && (
                <span className="ml-1 text-[10px] text-muted-foreground/50">(未配置)</span>
              )}
            </span>
          ))}
        </div>

        {/* Rows */}
        {categories.map((cat) => (
          <div
            key={cat.key}
            className="grid grid-cols-[1fr_80px_80px_80px] items-center gap-2 border-b border-border px-4 py-3 last:border-b-0"
          >
            <span className="text-sm text-foreground">{cat.label}</span>
            {channels.map((ch) => {
              const configured = isChannelConfigured(ch.key)
              const enabled = isChannelEnabled(cat.key, ch.key)
              return (
                <div key={ch.key} className="flex justify-center">
                  <Switch
                    checked={enabled}
                    disabled={!configured || saving}
                    onCheckedChange={(checked) => togglePreference(cat.key, ch.key, checked)}
                    aria-label={`${cat.label} - ${ch.label}`}
                  />
                </div>
              )
            })}
          </div>
        ))}
      </div>

      <p className="text-xs text-muted-foreground">
        未配置的渠道由管理员启用后才可选择。站内信始终可用。
      </p>
    </PageShell>
  )
}
