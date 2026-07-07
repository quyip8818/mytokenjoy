import { ProviderKeysPageShell, useProviderKeysPage } from '@/features/keys'

export default function ProviderKeysPage() {
  return <ProviderKeysPageShell {...useProviderKeysPage()} />
}
