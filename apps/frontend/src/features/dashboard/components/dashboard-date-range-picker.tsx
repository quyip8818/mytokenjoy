import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { CostPeriod } from '@/api/types'
import { COST_PERIOD_LABELS, DATE_RANGE_PRESETS } from '../lib/constants'

interface DashboardDateRangePickerProps {
  value: CostPeriod
  onChange: (value: string | null) => void
}

export function DashboardDateRangePicker({ value, onChange }: DashboardDateRangePickerProps) {
  return (
    <div className="flex items-center gap-1" role="group" aria-label="时间范围">
      {DATE_RANGE_PRESETS.map((preset) => (
        <Button
          key={preset}
          variant={value === preset ? 'secondary' : 'ghost'}
          size="sm"
          className={cn(
            'text-xs',
            value === preset && 'font-semibold',
          )}
          onClick={() => onChange(preset)}
          aria-pressed={value === preset}
        >
          {COST_PERIOD_LABELS[preset]}
        </Button>
      ))}
    </div>
  )
}
