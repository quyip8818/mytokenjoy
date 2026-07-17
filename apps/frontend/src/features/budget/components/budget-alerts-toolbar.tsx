import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

export type AlertTypeFilter = 'all' | 'team' | 'project'
export type AlertStatusFilter = 'all' | 'enabled' | 'disabled'

interface BudgetAlertsToolbarProps {
  typeFilter: AlertTypeFilter
  onTypeFilterChange: (value: AlertTypeFilter) => void
  statusFilter: AlertStatusFilter
  onStatusFilterChange: (value: AlertStatusFilter) => void
  search: string
  onSearchChange: (value: string) => void
}

const TYPE_TABS: { value: AlertTypeFilter; label: string }[] = [
  { value: 'all', label: '全部' },
  { value: 'team', label: '团队' },
  { value: 'project', label: '项目' },
]

export function BudgetAlertsToolbar({
  typeFilter,
  onTypeFilterChange,
  statusFilter,
  onStatusFilterChange,
  search,
  onSearchChange,
}: BudgetAlertsToolbarProps) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="flex items-center gap-1">
        {TYPE_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => onTypeFilterChange(tab.value)}
            className={cn(
              'rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-100',
              typeFilter === tab.value
                ? 'bg-muted text-foreground'
                : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <Select value={statusFilter} onValueChange={(v) => onStatusFilterChange(v as AlertStatusFilter)}>
          <SelectTrigger className="h-8 w-28 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="enabled">已启用</SelectItem>
            <SelectItem value="disabled">已禁用</SelectItem>
          </SelectContent>
        </Select>

        <div className="relative">
          <Search className="absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="搜索监控对象..."
            className="h-8 w-48 pl-8 text-sm"
          />
        </div>
      </div>
    </div>
  )
}
