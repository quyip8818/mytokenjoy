import { StructurePageShell, useStructurePage } from '@/features/org'

export default function StructurePage() {
  return <StructurePageShell {...useStructurePage()} />
}
