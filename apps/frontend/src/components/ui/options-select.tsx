import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

export interface OptionsSelectProps {
  value: string
  onValueChange: (value: string) => void
  options: Record<string, string>
  allValue?: string
  allLabel: string
  placeholder?: string
  className?: string
}

export function OptionsSelect({
  value,
  onValueChange,
  options,
  allValue = 'all',
  allLabel,
  placeholder,
  className,
}: OptionsSelectProps) {
  return (
    <Select value={value} onValueChange={(v) => onValueChange(v ?? allValue)}>
      <SelectTrigger className={cn('border-border/60 focus:ring-blue-500', className)}>
        <SelectValue placeholder={placeholder ?? allLabel} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={allValue}>{allLabel}</SelectItem>
        {Object.entries(options).map(([optionValue, label]) => (
          <SelectItem key={optionValue} value={optionValue}>
            {label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
