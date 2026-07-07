import { CallLogsPageShell, useAuditCallsPage } from '@/features/audit'

export default function CallLogsPage() {
  return <CallLogsPageShell {...useAuditCallsPage()} />
}
