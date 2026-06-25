import type { ReactNode } from 'react'
import { WorkflowFormLayout } from './workflow-form-layout'
import { WorkflowInfoBox } from './workflow-info-box'

interface WorkflowDetailLayoutProps {
  children: ReactNode
  sidebar: ReactNode
}

export function WorkflowDetailLayout({ children, sidebar }: WorkflowDetailLayoutProps) {
  return (
    <WorkflowFormLayout variant="full" className="grid grid-cols-5 gap-8">
      <div className="col-span-3 space-y-6">{children}</div>
      <WorkflowInfoBox fullWidth>{sidebar}</WorkflowInfoBox>
    </WorkflowFormLayout>
  )
}
