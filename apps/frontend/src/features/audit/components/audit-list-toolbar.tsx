import type { ReactNode } from 'react'
import { AuditKeywordInput, AuditMemberSelect, AuditToolbar } from '@/features/audit'
import { AuditDateRangePicker } from './audit-date-range-picker'

interface AuditListToolbarProps {
  datePreset: string
  onDatePresetChange: (value: string) => void
  memberId: string
  onMemberIdChange: (value: string) => void
  memberAllLabel: string
  memberOptions: Record<string, string>
  keyword: string
  onKeywordChange: (value: string) => void
  onExport: () => void
  children: ReactNode
}

export function AuditListToolbar({
  datePreset,
  onDatePresetChange,
  memberId,
  onMemberIdChange,
  memberAllLabel,
  memberOptions,
  keyword,
  onKeywordChange,
  onExport,
  children,
}: AuditListToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <AuditDateRangePicker value={datePreset} onChange={onDatePresetChange} />
      {children}
      <AuditMemberSelect
        value={memberId}
        onValueChange={onMemberIdChange}
        allLabel={memberAllLabel}
        options={memberOptions}
      />
      <AuditKeywordInput value={keyword} onChange={onKeywordChange} />
      <AuditToolbar onExport={onExport} />
    </div>
  )
}
