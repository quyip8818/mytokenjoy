import type { ReactNode } from 'react'
import { Label } from '@/components/ui/label'
import { WORKFLOW_FIELD_ERROR_CLASS, WORKFLOW_FORM_FIELD_CLASS } from '../constants'
import { cn } from '@/lib/utils'

interface WorkflowFormFieldProps {
  label: string
  error?: string
  children: ReactNode
  className?: string
}

export function WorkflowFormField({ label, error, children, className }: WorkflowFormFieldProps) {
  return (
    <div className={cn(WORKFLOW_FORM_FIELD_CLASS, className)}>
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
      {error ? <p className={WORKFLOW_FIELD_ERROR_CLASS}>{error}</p> : null}
    </div>
  )
}
