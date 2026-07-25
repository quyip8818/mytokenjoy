import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export interface FieldProps {
  label: string
  required?: boolean
  hint?: string
  className?: string
  children: ReactNode
}

export function Field({ label, required, hint, className, children }: FieldProps) {
  return (
    <div className={cn(className)}>
      <label className="mb-1 block text-sm font-medium">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </label>
      {children}
      {hint && <p className="mt-0.5 text-xs text-muted-foreground">{hint}</p>}
    </div>
  )
}
