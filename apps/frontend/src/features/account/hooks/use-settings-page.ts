import { useUrlTab } from '@/hooks/use-url-tab'
import { useAccountPage } from './use-account-page'
import { useLoginActivityPage } from './use-login-activity-page'
import { useNotificationsPage } from '@/features/notifications'

export type SettingsTab = 'account' | 'security' | 'notifications'

const SETTINGS_TABS = ['account', 'security', 'notifications'] as const

export function useSettingsPage() {
  const [activeTab, setActiveTab] = useUrlTab<SettingsTab>(SETTINGS_TABS, 'account')
  const accountPage = useAccountPage()
  const loginActivityPage = useLoginActivityPage()
  const notificationsPage = useNotificationsPage()

  return {
    activeTab,
    setActiveTab,
    accountPage,
    loginActivityPage,
    notificationsPage,
  }
}

export type SettingsPageState = ReturnType<typeof useSettingsPage>
