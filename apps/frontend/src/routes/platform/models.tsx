import { PlatformModelsPageShell, usePlatformModelsPage } from '@/features/platform'

export default function PlatformModelsPage() {
  return <PlatformModelsPageShell {...usePlatformModelsPage()} />
}
