import { PlatformKeysPageShell, usePlatformKeysPage } from '@/features/keys'

export default function PlatformKeysPage() {
  return <PlatformKeysPageShell {...usePlatformKeysPage()} />
}
