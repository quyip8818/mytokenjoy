import { OptionsSelect } from '@/components/ui/options-select'
import { OPERATION_ACTION_LABELS } from '@/features/audit/lib/labels'

interface OperationsLogFiltersProps {
  actionFilter: string
  onActionFilterChange: (value: string) => void
}

export function OperationsLogFilters({
  actionFilter,
  onActionFilterChange,
}: OperationsLogFiltersProps) {
  return (
    <OptionsSelect
      value={actionFilter}
      onValueChange={onActionFilterChange}
      options={OPERATION_ACTION_LABELS}
      allLabel="全部类型"
      className="w-40"
    />
  )
}
