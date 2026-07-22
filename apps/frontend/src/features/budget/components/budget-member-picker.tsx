import { useState } from 'react'
import type { Member } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { Users } from 'lucide-react'

interface BudgetMemberPickerProps {
  members: Member[]
  loading?: boolean
  selectedIds: string[]
  onChange: (ids: string[]) => void
  disabled?: boolean
}

export function BudgetMemberPicker({
  members,
  loading = false,
  selectedIds,
  onChange,
  disabled = false,
}: BudgetMemberPickerProps) {
  const [open, setOpen] = useState(false)

  const selectedMembers = members.filter((member) => selectedIds.includes(member.id))

  function toggle(id: string) {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((selectedId) => selectedId !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            'h-8 w-full justify-start gap-2 font-normal',
            !selectedIds.length && 'text-muted-foreground',
          )}
          disabled={disabled || loading}
          aria-label="选择关联成员"
        >
          <Users className="size-4 shrink-0" />
          {loading ? (
            <span>加载成员…</span>
          ) : selectedMembers.length === 0 ? (
            <span>{disabled ? '请先选择团队' : '选择成员…'}</span>
          ) : (
            <span className="flex flex-wrap gap-1">
              {selectedMembers.map((member) => (
                <Badge key={member.id} variant="outline" className="h-5 px-1.5 text-xs font-normal">
                  {member.alias}
                </Badge>
              ))}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-56 p-2" align="start">
        {members.length === 0 ? (
          <p className="px-1 py-2 text-xs text-muted-foreground">该团队暂无成员</p>
        ) : (
          <ul className="space-y-0.5" role="listbox" aria-multiselectable="true">
            {members.map((member) => {
              const checked = selectedIds.includes(member.id)
              return (
                <li key={member.id}>
                  <label className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted">
                    <Checkbox
                      checked={checked}
                      onCheckedChange={() => toggle(member.id)}
                      aria-label={member.alias}
                    />
                    <span className="flex-1 truncate">{member.alias}</span>
                    <span className="text-xs text-muted-foreground">{member.departmentName}</span>
                  </label>
                </li>
              )
            })}
          </ul>
        )}
      </PopoverContent>
    </Popover>
  )
}
