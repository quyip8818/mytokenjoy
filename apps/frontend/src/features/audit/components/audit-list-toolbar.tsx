import type { ReactNode } from 'react'
import { AuditDatePresetSelect } from './audit-date-preset-select'
import { AuditKeywordInput } from './audit-keyword-input'
import { AuditMemberSelect } from './audit-member-select'
import { AuditToolbar } from './audit-toolbar'

interface AuditListToolbarProps {
  datePreset: string
  onDatePresetChange: (value: string) => void
  memberId: string
  onMemberIdChange: (value: string) => void
  memberAllLabel: string
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
  keyword,
  onKeywordChange,
  onExport,
  children,
}: AuditListToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <AuditDatePresetSelect value={datePreset} onValueChange={onDatePresetChange} />
      {children}
      <AuditMemberSelect
        value={memberId}
        onValueChange={onMemberIdChange}
        allLabel={memberAllLabel}
      />
      <AuditKeywordInput value={keyword} onChange={onKeywordChange} />
      <AuditToolbar onExport={onExport} />
    </div>
  )
}
