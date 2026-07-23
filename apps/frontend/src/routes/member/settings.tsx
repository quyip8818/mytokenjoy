import { SettingsPageShell, useSettingsPage } from '@/features/account'

export default function MemberSettingsPage() {
  return <SettingsPageShell {...useSettingsPage()} />
}
