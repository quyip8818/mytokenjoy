import type { ComponentType } from 'react'
import { WorkflowDelegatePanel } from './components/workflow-delegate-panel'
import { useWorkflow } from './use-workflow'
import type { WorkflowComponentProps, WorkflowDefinition, WorkflowId, WorkflowLayer } from './types'
import { defineWorkflow } from './types'
import type { WorkflowPayloadMap } from './workflow-payloads'

interface DelegateWorkflowContext {
  onSaved: () => void
  onSetDirty: (dirty: boolean) => void
}

export function defineDelegateWorkflow<T extends WorkflowId, P extends object>(config: {
  id: T
  title: string
  defaultLayer: WorkflowLayer
  Child: ComponentType<P>
  mapProps: (payload: WorkflowPayloadMap[T], ctx: DelegateWorkflowContext) => P
}): WorkflowDefinition {
  function DelegateWorkflow({ entry, onClose, onSetDirty }: WorkflowComponentProps<T>) {
    const { closeAll } = useWorkflow()
    const onSuccess = (entry.payload as { onSuccess?: () => void }).onSuccess

    const onSaved = () => {
      onSuccess?.()
      closeAll()
    }

    const childProps = config.mapProps(entry.payload, { onSaved, onSetDirty })
    const Child = config.Child

    return (
      <WorkflowDelegatePanel title={config.title} onClose={onClose} onSetDirty={onSetDirty}>
        <Child {...(childProps as P)} />
      </WorkflowDelegatePanel>
    )
  }

  return defineWorkflow(DelegateWorkflow, {
    defaultLayer: config.defaultLayer,
    title: config.title,
  })
}
