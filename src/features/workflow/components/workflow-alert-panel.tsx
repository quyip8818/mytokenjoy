import type { LucideIcon } from 'lucide-react'
import { AlertTriangle } from 'lucide-react'
import { WorkflowFormLayout } from './workflow-form-layout'

interface WorkflowAlertPanelProps {
  title: string
  description: string
  icon?: LucideIcon
}

export function WorkflowAlertPanel({
  title,
  description,
  icon: Icon = AlertTriangle,
}: WorkflowAlertPanelProps) {
  return (
    <WorkflowFormLayout
      variant="wide"
      className="flex flex-col items-center justify-center py-12 text-center"
    >
      <div className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-50">
        <Icon className="h-7 w-7 text-amber-600" />
      </div>
      <div className="max-w-sm space-y-2">
        <p className="font-semibold">{title}</p>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </WorkflowFormLayout>
  )
}
