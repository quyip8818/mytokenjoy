import { useCallback } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { useAccountPage } from './use-account-page'
import { useLoginActivityPage } from './use-login-activity-page'
import { useNotificationsPage } from '@/features/notifications'

export type SettingsTab = 'account' | 'security' | 'notifications'

export function useSettingsPage() {
  const { tab } = useSearch({ strict: false }) as { tab?: SettingsTab }
  const navigate = useNavigate()

  const activeTab: SettingsTab = tab ?? 'account'

  const setActiveTab = useCallback(
    (newTab: SettingsTab) => {
      void navigate({
        to: '/me/settings',
        search: { tab: newTab },
        replace: true,
      })
    },
    [navigate],
  )

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
