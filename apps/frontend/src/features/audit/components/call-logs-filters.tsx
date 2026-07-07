import { OptionsSelect } from '@/components/ui/options-select'
import { CALL_LOG_STATUS_LABELS } from '@/features/audit'

interface CallLogsFiltersProps {
  statusFilter: string
  modelFilter: string
  modelOptions: Record<string, string>
  onStatusChange: (value: string) => void
  onModelChange: (value: string) => void
}

export function CallLogsFilters({
  statusFilter,
  modelFilter,
  modelOptions,
  onStatusChange,
  onModelChange,
}: CallLogsFiltersProps) {
  return (
    <>
      <OptionsSelect
        value={statusFilter}
        onValueChange={onStatusChange}
        options={CALL_LOG_STATUS_LABELS}
        allLabel="全部状态"
        className="w-32"
      />
      <OptionsSelect
        value={modelFilter}
        onValueChange={onModelChange}
        options={modelOptions}
        allLabel="全部模型"
        className="w-40"
      />
    </>
  )
}
