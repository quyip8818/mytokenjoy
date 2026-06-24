import { SyncConfigPanel } from '@/components/org/sync-config'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome } from '../components/workflow-panel-chrome'
import { useWorkflow } from '../use-workflow'

export function SyncConfigWorkflow({ entry, onClose, onSetDirty }: WorkflowComponentProps) {
  const { closeAll } = useWorkflow()
  const onTriggerSync = entry.payload.onTriggerSync as (() => void) | undefined
  const triggeringSync = (entry.payload.triggeringSync as boolean) ?? false
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined

  return (
    <WorkflowPanelChrome title="同步策略" onClose={onClose}>
      <div onChange={() => onSetDirty(true)}>
        <SyncConfigPanel
          onTriggerSync={onTriggerSync ?? (() => {})}
          triggeringSyc={triggeringSync}
          onSaved={() => {
            onSuccess?.()
            closeAll()
          }}
        />
      </div>
    </WorkflowPanelChrome>
  )
}
