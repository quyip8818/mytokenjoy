import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AUDIT_DATE_PRESET, AUDIT_DATE_PRESET_LABELS } from '@/features/audit'

interface AuditDatePresetSelectProps {
  value: string
  onValueChange: (value: string) => void
}

export function AuditDatePresetSelect({ value, onValueChange }: AuditDatePresetSelectProps) {
  return (
    <Select value={value} onValueChange={(v) => onValueChange(v ?? AUDIT_DATE_PRESET.ALL)}>
      <SelectTrigger className="w-32 border-border/60 focus:ring-blue-500">
        <SelectValue placeholder="时间范围" />
      </SelectTrigger>
      <SelectContent>
        {Object.entries(AUDIT_DATE_PRESET_LABELS).map(([preset, label]) => (
          <SelectItem key={preset} value={preset}>
            {label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
