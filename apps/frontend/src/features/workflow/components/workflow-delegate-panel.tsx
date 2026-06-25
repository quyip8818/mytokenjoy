import type { ReactElement } from 'react'
import { WorkflowPanelChrome } from './workflow-panel-chrome'
import { WorkflowFormLayout } from './workflow-form-layout'

interface WorkflowDelegatePanelProps {
  title: string
  onClose: () => void
  onSetDirty: (dirty: boolean) => void
  children: ReactElement
}

export function WorkflowDelegatePanel({
  title,
  onClose,
  onSetDirty,
  children,
}: WorkflowDelegatePanelProps) {
  return (
    <WorkflowPanelChrome title={title} onClose={onClose}>
      <WorkflowFormLayout as="div" onChange={() => onSetDirty(true)}>
        {children}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
