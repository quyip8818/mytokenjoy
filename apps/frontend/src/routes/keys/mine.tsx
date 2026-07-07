import { MyKeysPageShell, useMyKeysPage } from '@/features/keys'

export default function MyKeysPage() {
  return <MyKeysPageShell {...useMyKeysPage()} />
}
