import { PlatformCompaniesPageShell, usePlatformCompaniesPage } from '@/features/platform'

export default function PlatformCompaniesPage() {
  return <PlatformCompaniesPageShell {...usePlatformCompaniesPage()} />
}
