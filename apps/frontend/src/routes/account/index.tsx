import { AccountPageShell, useAccountPage } from '@/features/account'

export default function AccountPage() {
  return <AccountPageShell {...useAccountPage()} />
}
