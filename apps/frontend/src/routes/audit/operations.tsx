import { OperationsLogPageShell, useAuditOperationsPage } from '@/features/audit'

export default function OperationLogsPage() {
  return <OperationsLogPageShell {...useAuditOperationsPage()} />
}
