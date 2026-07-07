import { MyKeysAdminPageShell, useMyKeysPage } from '@/features/keys'

export default function MyKeysPage() {
  return <MyKeysAdminPageShell {...useMyKeysPage()} />
}
