import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search, UserPlus, Plus } from 'lucide-react'

interface MemberToolbarProps {
  keyword: string
  onKeywordChange: (keyword: string) => void
  onInvite: () => void
  onAdd: () => void
}

export function MemberToolbar({ keyword, onKeywordChange, onInvite, onAdd }: MemberToolbarProps) {
  return (
    <div className="flex items-center gap-3">
      <div className="relative w-56">
        <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="搜索成员..."
          className="h-8 pl-8 text-sm"
          value={keyword}
          onChange={(e) => onKeywordChange(e.target.value)}
        />
      </div>
      <div className="flex-1" />
      <Button variant="outline" size="sm" onClick={onInvite}>
        <UserPlus className="size-3.5" />邀请成员
      </Button>
      <Button size="sm" onClick={onAdd}>
        <Plus className="size-3.5" />添加成员
      </Button>
    </div>
  )
}
