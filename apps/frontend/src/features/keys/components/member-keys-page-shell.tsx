import type { useMyKeysPage } from '@/features/keys'
import { MyKeysPageShell } from './my-keys-page-shell'

type MemberKeysPageShellProps = ReturnType<typeof useMyKeysPage>

export function MemberKeysPageShell(props: MemberKeysPageShellProps) {
  return (
    <MyKeysPageShell
      {...props}
      memberPortal
      description={<h1 className="text-sm font-semibold">我的 Key</h1>}
    />
  )
}
