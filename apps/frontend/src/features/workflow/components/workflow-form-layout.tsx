import type { FormHTMLAttributes, HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils'

type WorkflowFormLayoutVariant = 'narrow' | 'wide' | 'full'

const VARIANT_CLASSES: Record<WorkflowFormLayoutVariant, string> = {
  narrow: 'max-w-md space-y-4',
  wide: 'max-w-lg space-y-5',
  full: 'space-y-4',
}

interface WorkflowFormLayoutBaseProps {
  variant?: WorkflowFormLayoutVariant
  className?: string
  children: ReactNode
}

type WorkflowFormLayoutProps = WorkflowFormLayoutBaseProps &
  (
    | ({ as?: 'div' } & HTMLAttributes<HTMLDivElement>)
    | ({ as: 'form' } & FormHTMLAttributes<HTMLFormElement>)
  )

export function WorkflowFormLayout({
  variant = 'narrow',
  className,
  children,
  as = 'div',
  ...props
}: WorkflowFormLayoutProps) {
  const layoutClass = cn(VARIANT_CLASSES[variant], className)

  if (as === 'form') {
    const formProps = props as FormHTMLAttributes<HTMLFormElement>
    return (
      <form className={layoutClass} {...formProps}>
        {children}
      </form>
    )
  }

  const divProps = props as HTMLAttributes<HTMLDivElement>
  return (
    <div className={layoutClass} {...divProps}>
      {children}
    </div>
  )
}
