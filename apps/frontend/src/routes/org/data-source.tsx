import { DataSourcePageShell, useDataSourcePage } from '@/features/org'

export default function DataSourcePage() {
  return <DataSourcePageShell {...useDataSourcePage()} />
}
