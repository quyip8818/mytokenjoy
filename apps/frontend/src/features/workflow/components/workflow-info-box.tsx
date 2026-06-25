import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import {
  WORKFLOW_DROPZONE_CLASS,
  WORKFLOW_INFO_BOX_CLASS,
  WORKFLOW_INFO_BOX_CODE_CLASS,
} from '../constants'

type WorkflowInfoBoxVariant = 'default' | 'code' | 'dropzone'

interface WorkflowInfoBoxProps {
  variant?: WorkflowInfoBoxVariant
  fullWidth?: boolean
  className?: string
  children: ReactNode
  onClick?: () => void
}

export function WorkflowInfoBox({
  variant = 'default',
  fullWidth,
  className,
  children,
  onClick,
}: WorkflowInfoBoxProps) {
  const baseClass =
    variant === 'code'
      ? WORKFLOW_INFO_BOX_CODE_CLASS
      : variant === 'dropzone'
        ? WORKFLOW_DROPZONE_CLASS
        : WORKFLOW_INFO_BOX_CLASS

  const sharedClass = cn(baseClass, fullWidth && 'col-span-2', className)

  if (variant === 'dropzone') {
    return (
      <button type="button" className={sharedClass} onClick={onClick}>
        {children}
      </button>
    )
  }

  if (variant === 'code') {
    return <div className={cn(sharedClass, 'flex items-center gap-2')}>{children}</div>
  }

  return <div className={sharedClass}>{children}</div>
}
