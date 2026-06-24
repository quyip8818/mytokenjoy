import type { Platform } from '@/api/types'
import { CredentialForm } from '@/components/org/credential-form'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome } from '../components/workflow-panel-chrome'
import { useWorkflow } from '../use-workflow'

export function CredentialFormWorkflow({ entry, onClose, onSetDirty }: WorkflowComponentProps) {
  const { closeAll } = useWorkflow()
  const connected = (entry.payload.connected as boolean) ?? false
  const currentPlatform = (entry.payload.currentPlatform as Platform | null) ?? null
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined

  return (
    <WorkflowPanelChrome title="配置凭证" onClose={onClose}>
      <div onChange={() => onSetDirty(true)}>
        <CredentialForm
          connected={connected}
          currentPlatform={currentPlatform}
          onSaved={() => {
            onSuccess?.()
            closeAll()
          }}
        />
      </div>
    </WorkflowPanelChrome>
  )
}
