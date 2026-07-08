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
  onStepClick?: (index: number) => void
}

export function Stepper({ steps, currentStep, completedSteps, onStepClick }: StepperProps) {
  return (
    <nav aria-label="Progress" className="w-full">
      <ol className="flex items-center">
        {steps.map((step, index) => {
          const isCompleted = completedSteps.includes(index)
          const isCurrent = index === currentStep
          const isLast = index === steps.length - 1
          const isClickable = Boolean(onStepClick) && isCompleted && !isCurrent

          return (
            <li key={step.title} className={cn('relative flex items-center', !isLast && 'flex-1')}>
              <button
                type="button"
                disabled={!isClickable}
                onClick={() => isClickable && onStepClick?.(index)}
                aria-current={isCurrent ? 'step' : undefined}
                title={isClickable ? `返回「${step.title}」` : undefined}
                className={cn(
                  'flex items-center gap-3 rounded-md text-left outline-none',
                  'focus-visible:ring-2 focus-visible:ring-primary/20',
                  isClickable ? 'cursor-pointer' : 'cursor-default',
                )}
              >
                <div
                  className={cn(
                    'flex size-8 shrink-0 items-center justify-center rounded-full border-2 text-xs font-semibold transition-colors duration-150',
                    isCompleted && 'border-primary bg-primary text-primary-foreground',
                    isCurrent && !isCompleted && 'border-primary text-primary bg-primary/5',
                    !isCurrent && !isCompleted && 'border-border text-muted-foreground bg-card',
                  )}
                >
                  {isCompleted ? <CheckIcon className="size-4" /> : <span>{index + 1}</span>}
                </div>
                <div className="hidden sm:block">
                  <p
                    className={cn(
                      'text-sm font-medium leading-tight transition-colors duration-150',
                      isCurrent || isCompleted ? 'text-foreground' : 'text-muted-foreground',
                      isClickable && 'group-hover:text-primary',
                    )}
                  >
                    {step.title}
                  </p>
                  {step.description && (
                    <p className="text-xs text-muted-foreground mt-0.5">{step.description}</p>
                  )}
                </div>
              </button>
              {!isLast && (
                <div
                  className={cn(
                    'mx-4 h-0.5 flex-1 rounded-full transition-colors duration-150',
                    isCompleted ? 'bg-primary' : 'bg-border',
                  )}
                />
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
