import { cn } from '@/lib/utils'

interface WorkflowStepperProps {
  steps: string[]
  current: number
}

export function WorkflowStepper({ steps, current }: WorkflowStepperProps) {
  return (
    <div className="flex items-center gap-2 mb-6">
      {steps.map((label, index) => {
        const stepNum = index + 1
        const active = stepNum === current
        const done = stepNum < current
        return (
          <div key={label} className="flex items-center gap-2">
            {index > 0 && <div className="h-px w-8 bg-border" />}
            <div className="flex items-center gap-2">
              <div
                className={cn(
                  'flex h-7 w-7 items-center justify-center rounded-full text-xs font-semibold',
                  active && 'bg-indigo-600 text-white',
                  done && 'bg-indigo-100 text-indigo-700',
                  !active && !done && 'bg-slate-100 text-muted-foreground',
                )}
              >
                {stepNum}
              </div>
              <span
                className={cn(
                  'text-sm',
                  active ? 'font-medium text-foreground' : 'text-muted-foreground',
                )}
              >
                {label}
              </span>
            </div>
          </div>
        )
      })}
    </div>
  )
}
