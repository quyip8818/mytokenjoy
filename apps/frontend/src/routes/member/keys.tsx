import { MemberKeysPageShell, useMyKeysPage } from '@/features/keys'

export default function MemberKeysPage() {
  return <MemberKeysPageShell {...useMyKeysPage()} />
}
