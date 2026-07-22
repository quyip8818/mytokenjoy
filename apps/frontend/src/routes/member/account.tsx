import { AccountPageShell, useAccountPage } from '@/features/account'

export default function MemberAccountPage() {
  return <AccountPageShell {...useAccountPage()} />
}
