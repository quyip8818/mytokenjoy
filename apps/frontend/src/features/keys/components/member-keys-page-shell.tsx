import { MyKeysPageShell } from './my-keys-page-shell'

export function MemberKeysPageShell() {
  return (
    <MyKeysPageShell
      memberPortal
      description={<h1 className="text-sm font-semibold">我的 Key</h1>}
    />
  )
}
