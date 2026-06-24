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
      <div className="flex items-center gap-2">
        {steps.map((step, index) => (
          <div key={step.id} className="flex flex-1 items-center gap-2">
            {index > 0 && (
              <div className={cn('h-0.5 flex-1', step.done ? 'bg-blue-400' : 'bg-border')} />
            )}
            <div className="flex flex-col items-center gap-1 min-w-0">
              <div
                className={cn(
                  'flex h-8 w-8 items-center justify-center rounded-full border-2 text-xs font-medium shrink-0',
                  step.done
                    ? 'border-blue-600 bg-blue-600 text-white'
                    : 'border-border bg-background text-muted-foreground',
                )}
              >
                {step.done ? <Check className="h-4 w-4" /> : index + 1}
              </div>
              <span
                className={cn(
                  'text-[11px] text-center leading-tight',
                  step.done ? 'text-blue-700 font-medium' : 'text-muted-foreground',
                )}
              >
                {step.label}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
