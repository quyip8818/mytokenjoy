import { RolesPageShell, useRolesPage } from '@/features/org'

export default function RolesPage() {
  return <RolesPageShell {...useRolesPage()} />
}
