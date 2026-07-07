import { MemberDashboardPageShell, useMemberDashboardPage } from '@/features/member'

export default function MemberDashboardPage() {
  return <MemberDashboardPageShell {...useMemberDashboardPage()} />
}
