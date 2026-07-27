import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export interface SplitPanelProps {
  master: ReactNode
  detail: ReactNode
  masterWidth?: number
  className?: string
}

export function SplitPanel({ master, detail, masterWidth = 280, className }: SplitPanelProps) {
  return (
    <div
      className={cn(
        'flex min-h-0 flex-1 overflow-hidden rounded-lg border border-border bg-card shadow-xs',
        className,
      )}
    >
      <div
        className="shrink-0 overflow-y-auto border-r border-border"
        style={{ width: masterWidth }}
      >
        {master}
      </div>
      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">{detail}</div>
    </div>
  )
}
