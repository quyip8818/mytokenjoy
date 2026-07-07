import { cn } from '@/lib/utils'
import { CheckIcon } from 'lucide-react'

interface Step {
  title: string
  description?: string
}

interface StepperProps {
  steps: Step[]
  currentStep: number
  completedSteps: number[]
}

export function Stepper({ steps, currentStep, completedSteps }: StepperProps) {
  return (
    <nav aria-label="Progress" className="w-full">
      <ol className="flex items-center">
        {steps.map((step, index) => {
          const isCompleted = completedSteps.includes(index)
          const isCurrent = index === currentStep
          const isLast = index === steps.length - 1

          return (
            <li key={step.title} className={cn('relative flex items-center', !isLast && 'flex-1')}>
              <div className="flex items-center gap-3">
                <div
                  className={cn(
                    'flex size-7 shrink-0 items-center justify-center rounded-full border text-xs font-medium',
                    isCompleted && 'border-primary bg-primary text-white',
                    isCurrent && !isCompleted && 'border-primary text-primary bg-card',
                    !isCurrent && !isCompleted && 'border-border text-muted-foreground bg-card',
                  )}
                >
                  {isCompleted ? <CheckIcon className="size-3.5" /> : <span>{index + 1}</span>}
                </div>
                <div className="hidden sm:block">
                  <p
                    className={cn(
                      'text-sm font-medium leading-tight',
                      isCurrent || isCompleted ? 'text-foreground' : 'text-muted-foreground',
                    )}
                  >
                    {step.title}
                  </p>
                  {step.description && (
                    <p className="text-xs text-muted-foreground mt-0.5">{step.description}</p>
                  )}
                </div>
              </div>
              {!isLast && (
                <div className={cn('mx-4 h-px flex-1', isCompleted ? 'bg-primary' : 'bg-border')} />
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
