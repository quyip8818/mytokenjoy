import { SyncConfigPanel } from '@/components/org/sync-config'
import type { WorkflowComponentProps } from '../types'
import { WorkflowDelegatePanel } from '../components/workflow-delegate-panel'
import { useWorkflow } from '../use-workflow'

export function SyncConfigWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'sync-config'>) {
  const { closeAll } = useWorkflow()
  const onTriggerSync = entry.payload.onTriggerSync as (() => void) | undefined
  const triggeringSync = (entry.payload.triggeringSync as boolean) ?? false
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined

  return (
    <WorkflowDelegatePanel title="同步策略" onClose={onClose} onSetDirty={onSetDirty}>
      <SyncConfigPanel
        onTriggerSync={onTriggerSync ?? (() => {})}
        triggeringSync={triggeringSync}
        onSaved={() => {
          onSuccess?.()
          closeAll()
        }}
      />
    </WorkflowDelegatePanel>
  )
}
