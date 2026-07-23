import { useState } from 'react'
import { useAccountPage } from './use-account-page'
import { useLoginActivityPage } from './use-login-activity-page'
import { useNotificationsPage } from '@/features/notifications'

export type SettingsTab = 'account' | 'security' | 'notifications'

export function useSettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('account')
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
