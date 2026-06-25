import { Fragment } from 'react'
import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface InitProgressStep {
  id: string
  label: string
  done: boolean
}

interface DataSourceInitProgressProps {
  connected: boolean
  imported: boolean
}

export function DataSourceInitProgress({ connected, imported }: DataSourceInitProgressProps) {
  const steps: InitProgressStep[] = [
    { id: 'connect', label: '连接数据源', done: connected },
    { id: 'import', label: '全量导入组织', done: imported },
    { id: 'ready', label: '进入组织架构', done: imported },
  ]

  const doneCount = steps.filter((s) => s.done).length

  return (
    <div className="space-y-3 rounded-lg border border-border/50 bg-muted/30 px-4 py-4 shadow-card">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium">平台初始化进度</p>
        <span className="text-xs text-muted-foreground">
          {doneCount}/{steps.length}
        </span>
      </div>
      <div className="flex w-full items-start">
        {steps.map((step, index) => (
          <Fragment key={step.id}>
            {index > 0 && (
              <div
                className={cn(
                  'mt-4 h-0.5 min-w-6 flex-1',
                  steps[index - 1].done ? 'bg-blue-400' : 'bg-border',
                )}
              />
            )}
            <div className="flex max-w-[7rem] shrink-0 flex-col items-center gap-1">
              <div
                className={cn(
                  'flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2 text-xs font-medium',
                  step.done
                    ? 'border-blue-600 bg-blue-600 text-white'
                    : 'border-border bg-background text-muted-foreground',
                )}
              >
                {step.done ? <Check className="h-4 w-4" /> : index + 1}
              </div>
              <span
                className={cn(
                  'text-center text-[11px] leading-tight',
                  step.done ? 'font-medium text-blue-700' : 'text-muted-foreground',
                )}
              >
                {step.label}
              </span>
            </div>
          </Fragment>
        ))}
      </div>
    </div>
  )
}
