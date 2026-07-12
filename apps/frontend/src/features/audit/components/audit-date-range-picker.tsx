import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { AUDIT_DATE_PRESET_LABELS } from '../lib/constants'

interface AuditDateRangePickerProps {
  value: string
  onChange: (value: string) => void
}

const PRESETS = Object.keys(AUDIT_DATE_PRESET_LABELS)

export function AuditDateRangePicker({ value, onChange }: AuditDateRangePickerProps) {
  return (
    <div className="flex items-center gap-1" role="group" aria-label="时间范围">
      {PRESETS.map((preset) => (
        <Button
          key={preset}
          variant={value === preset ? 'secondary' : 'ghost'}
          size="sm"
          className={cn('text-xs', value === preset && 'font-semibold')}
          onClick={() => onChange(preset)}
          aria-pressed={value === preset}
        >
          {AUDIT_DATE_PRESET_LABELS[preset]}
        </Button>
      ))}
    </div>
  )
}
