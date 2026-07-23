import { cn } from '@/lib/utils'
import { AccountPageShell } from './account-page-shell'
import { LoginActivityPageShell } from './login-activity-page-shell'
import { NotificationsPageShell } from '@/features/notifications'
import type { SettingsPageState, SettingsTab } from '../hooks/use-settings-page'

const TABS: { key: SettingsTab; label: string }[] = [
  { key: 'account', label: '基本信息' },
  { key: 'security', label: '安全' },
  { key: 'notifications', label: '通知' },
]

export function SettingsPageShell({
  activeTab,
  setActiveTab,
  accountPage,
  loginActivityPage,
  notificationsPage,
}: SettingsPageState) {
  return (
    <div className="space-y-4">
      <nav className="flex gap-1 border-b border-border" aria-label="设置">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            type="button"
            onClick={() => setActiveTab(tab.key)}
            className={cn(
              'px-4 py-2 text-sm font-medium transition-colors',
              activeTab === tab.key
                ? 'border-b-2 border-primary text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
            aria-current={activeTab === tab.key ? 'page' : undefined}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {activeTab === 'account' && <AccountPageShell {...accountPage} />}
      {activeTab === 'security' && <LoginActivityPageShell {...loginActivityPage} />}
      {activeTab === 'notifications' && <NotificationsPageShell {...notificationsPage} />}
    </div>
  )
}
