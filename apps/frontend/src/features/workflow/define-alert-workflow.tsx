import type { ComponentType } from 'react'
import { WorkflowPanelChrome, WorkflowPanelFooter } from './components/workflow-panel-chrome'
import { WorkflowAlertPanel } from './components/workflow-alert-panel'
import type { WorkflowComponentProps, WorkflowDefinition, WorkflowId, WorkflowLayer } from './types'
import { defineWorkflow } from './types'
import type { WorkflowPayloadMap } from './workflow-payloads'

export function defineAlertWorkflow<T extends WorkflowId>(config: {
  id: T
  title: string
  defaultLayer: WorkflowLayer
  alert: (payload: WorkflowPayloadMap[T]) => { title: string; description: string }
  primaryLabel?: string
}): WorkflowDefinition {
  function AlertWorkflow({ entry, onPop, onClose }: WorkflowComponentProps<T>) {
    const { title, description } = config.alert(entry.payload)

    return (
      <WorkflowPanelChrome
        title={config.title}
        showBack
        onBack={onPop}
        onClose={onClose}
        footer={
          <WorkflowPanelFooter primaryLabel={config.primaryLabel ?? '知道了'} onPrimary={onPop} />
        }
      >
        <WorkflowAlertPanel title={title} description={description} />
      </WorkflowPanelChrome>
    )
  }

  return defineWorkflow(AlertWorkflow as ComponentType<WorkflowComponentProps<T>>, {
    defaultLayer: config.defaultLayer,
    title: config.title,
  })
}
