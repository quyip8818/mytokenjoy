import { LoginActivityPageShell, useLoginActivityPage } from '@/features/account'

export default function MemberLoginActivityPage() {
  return <LoginActivityPageShell {...useLoginActivityPage()} />
}
