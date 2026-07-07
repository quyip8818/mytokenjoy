import { OptionsSelect } from '@/components/ui/options-select'

interface AuditMemberSelectProps {
  value: string
  onValueChange: (value: string) => void
  allLabel: string
  options: Record<string, string>
  placeholder?: string
  className?: string
}

export function AuditMemberSelect({
  value,
  onValueChange,
  allLabel,
  options,
  placeholder,
  className = 'w-36',
}: AuditMemberSelectProps) {
  return (
    <OptionsSelect
      value={value}
      onValueChange={onValueChange}
      options={options}
      allLabel={allLabel}
      placeholder={placeholder ?? allLabel}
      className={className}
    />
  )
}
