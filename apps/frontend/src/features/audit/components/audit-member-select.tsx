import { OptionsSelect } from '@/components/ui/options-select'
import { useAuditMemberOptions } from '../hooks/use-audit-member-options'

interface AuditMemberSelectProps {
  value: string
  onValueChange: (value: string) => void
  allLabel: string
  placeholder?: string
  className?: string
}

export function AuditMemberSelect({
  value,
  onValueChange,
  allLabel,
  placeholder,
  className = 'w-36',
}: AuditMemberSelectProps) {
  const { members } = useAuditMemberOptions()
  const memberOptions = Object.fromEntries(members.map((member) => [member.id, member.name]))

  return (
    <OptionsSelect
      value={value}
      onValueChange={onValueChange}
      options={memberOptions}
      allLabel={allLabel}
      placeholder={placeholder ?? allLabel}
      className={className}
    />
  )
}
