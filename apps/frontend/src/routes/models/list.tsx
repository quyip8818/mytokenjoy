import { ModelListPageShell, useModelListPage } from '@/features/models'

export default function ModelListPage() {
  return <ModelListPageShell {...useModelListPage()} />
}
