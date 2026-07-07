import { ApprovalPageShell, useApprovalPage } from '@/features/keys'

export default function ApprovalPage() {
  return <ApprovalPageShell {...useApprovalPage()} />
}
