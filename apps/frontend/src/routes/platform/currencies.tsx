import { PlatformCurrenciesPageShell, usePlatformCurrenciesPage } from '@/features/platform'

export default function PlatformCurrenciesPage() {
  return <PlatformCurrenciesPageShell {...usePlatformCurrenciesPage()} />
}
