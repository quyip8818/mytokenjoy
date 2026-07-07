import { useInjectedApis } from '@/api/use-apis'
import { CredentialForm, SyncConfigPanel } from '@/features/org'
import type { ComponentProps } from 'react'

type CredentialBridgeProps = Omit<ComponentProps<typeof CredentialForm>, 'dataSourceApi'>
type SyncConfigBridgeProps = Omit<ComponentProps<typeof SyncConfigPanel>, 'syncApi'>

export function OrgCredentialFormBridge(props: CredentialBridgeProps) {
  const apis = useInjectedApis()
  return <CredentialForm {...props} dataSourceApi={apis.dataSourceApi} />
}

export function OrgSyncConfigBridge(props: SyncConfigBridgeProps) {
  const apis = useInjectedApis()
  return <SyncConfigPanel {...props} syncApi={apis.syncApi} />
}
