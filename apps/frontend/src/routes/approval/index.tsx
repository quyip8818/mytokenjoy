import { ApprovalPageShell, useApprovalPage } from '@/features/approval'

export default function ApprovalPage() {
  return <ApprovalPageShell {...useApprovalPage()} />
}
