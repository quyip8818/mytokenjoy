import type { Platform } from '@/api/types'
import { CredentialForm } from '@/components/org/credential-form'
import type { WorkflowComponentProps } from '../types'
import { WorkflowDelegatePanel } from '../components/workflow-delegate-panel'
import { useWorkflow } from '../use-workflow'

export function CredentialFormWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'credential-form'>) {
  const { closeAll } = useWorkflow()
  const connected = (entry.payload.connected as boolean) ?? false
  const currentPlatform = (entry.payload.currentPlatform as Platform | null) ?? null
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined

  return (
    <WorkflowDelegatePanel title="配置凭证" onClose={onClose} onSetDirty={onSetDirty}>
      <CredentialForm
        connected={connected}
        currentPlatform={currentPlatform}
        onSaved={() => {
          onSuccess?.()
          closeAll()
        }}
      />
    </WorkflowDelegatePanel>
  )
}
