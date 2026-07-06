import { useState, useMemo } from 'react'
import { mockMembers } from '@/mocks/data'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { Users } from 'lucide-react'

interface BudgetMemberPickerProps {
  departmentId: string
  selectedIds: string[]
  onChange: (ids: string[]) => void
}

export function BudgetMemberPicker({
  departmentId,
  selectedIds,
  onChange,
}: BudgetMemberPickerProps) {
  const [open, setOpen] = useState(false)

  const members = useMemo(
    () => (departmentId ? mockMembers.filter((m) => m.departmentId === departmentId) : []),
    [departmentId]
  )

  const selectedMembers = useMemo(
    () => mockMembers.filter((m) => selectedIds.includes(m.id)),
    [selectedIds]
  )

  function toggle(id: string) {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((s) => s !== id))
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
            !selectedIds.length && 'text-muted-foreground'
          )}
          disabled={!departmentId}
          aria-label="选择关联成员"
        >
          <Users className="size-4 shrink-0" />
          {selectedMembers.length === 0 ? (
            <span>{departmentId ? '选择成员…' : '请先选择团队'}</span>
          ) : (
            <span className="flex flex-wrap gap-1">
              {selectedMembers.map((m) => (
                <Badge
                  key={m.id}
                  variant="outline"
                  className="h-5 px-1.5 text-xs font-normal"
                >
                  {m.name}
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
            {members.map((m) => {
              const checked = selectedIds.includes(m.id)
              return (
                <li key={m.id}>
                  <label
                    className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted"
                  >
                    <Checkbox
                      checked={checked}
                      onCheckedChange={() => toggle(m.id)}
                      aria-label={m.name}
                    />
                    <span className="flex-1 truncate">{m.name}</span>
                    <span className="text-xs text-muted-foreground">{m.jobTitle}</span>
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
