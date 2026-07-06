import type { ReactNode } from 'react'
import { WORKFLOW_LIST_ITEM_CLASS, WORKFLOW_LIST_ITEM_SELECTED_CLASS } from '../constants'
import { cn } from '@/lib/utils'

interface WorkflowListItemProps {
  selected?: boolean
  onClick: () => void
  children: ReactNode
  className?: string
}

export function WorkflowListItem({
  selected = false,
  onClick,
  children,
  className,
}: WorkflowListItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'w-full px-4 py-3 text-left text-sm',
        WORKFLOW_LIST_ITEM_CLASS,
        selected && WORKFLOW_LIST_ITEM_SELECTED_CLASS,
        className,
      )}
    >
      {children}
    </button>
  )
}

interface WorkflowScrollListProps {
  children: ReactNode
  className?: string
}

export function WorkflowScrollList({ children, className }: WorkflowScrollListProps) {
  return <div className={cn('max-h-[60vh] space-y-1 overflow-y-auto', className)}>{children}</div>
}
