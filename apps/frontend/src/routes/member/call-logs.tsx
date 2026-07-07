import { MemberCallLogsPageShell, useMemberCallLogsPage } from '@/features/member'

export default function MemberCallLogsPage() {
  return <MemberCallLogsPageShell {...useMemberCallLogsPage()} />
}
