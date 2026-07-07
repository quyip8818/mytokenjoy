import { ModelRoutingPageShell, useModelRoutingPage } from '@/features/models'

export default function ModelRoutingPage() {
  return <ModelRoutingPageShell {...useModelRoutingPage()} />
}
