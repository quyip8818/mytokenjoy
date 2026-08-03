import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { Search, Plus } from 'lucide-react'

interface MemberToolbarProps {
  keyword: string
  onKeywordChange: (keyword: string) => void
  onSearch: () => void
  onAdd: () => void
}

export function MemberToolbar({ keyword, onKeywordChange, onSearch, onAdd }: MemberToolbarProps) {
  return (
    <div className="flex items-center gap-3">
      <div className="relative w-56">
        <Input
          type="text"
          placeholder="搜索成员..."
          className="h-8 pr-8 text-sm"
          value={keyword}
          onChange={(e) => onKeywordChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') onSearch()
          }}
        />
        <button
          type="button"
          className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:text-foreground"
          onClick={onSearch}
        >
          <Search className="size-3.5" />
        </button>
      </div>
      <div className="flex-1" />
      <PermissionGate permission={PERMISSION.ORG_MANAGE}>
        <Button size="sm" onClick={onAdd}>
          <Plus className="size-3.5" />
          添加成员
        </Button>
      </PermissionGate>
    </div>
  )
}
