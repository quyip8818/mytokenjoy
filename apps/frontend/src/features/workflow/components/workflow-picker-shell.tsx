import type { ReactNode } from 'react'
import { WorkflowPanelChrome, WorkflowPanelFooter } from './workflow-panel-chrome'

interface WorkflowPickerShellProps {
  title: string
  onPop: () => void
  onClose: () => void
  primaryLabel?: string
  primaryDisabled?: boolean
  onConfirm: () => void
  children: ReactNode
}

export function WorkflowPickerShell({
  title,
  onPop,
  onClose,
  primaryLabel = '确认',
  primaryDisabled = false,
  onConfirm,
  children,
}: WorkflowPickerShellProps) {
  return (
    <WorkflowPanelChrome
      title={title}
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel={primaryLabel}
          onPrimary={onConfirm}
          primaryDisabled={primaryDisabled}
        />
      }
    >
      {children}
    </WorkflowPanelChrome>
  )
}
