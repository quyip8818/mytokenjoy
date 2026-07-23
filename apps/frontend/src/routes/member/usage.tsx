import { MemberCallLogsPageShell, useMemberCallLogsPage } from '@/features/member'

export default function MemberUsagePage() {
  return <MemberCallLogsPageShell {...useMemberCallLogsPage()} />
}
