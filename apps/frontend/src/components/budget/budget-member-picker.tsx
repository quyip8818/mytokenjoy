import { useEffect, useMemo, useState } from 'react'
import { useInjectedApis } from '@/api/use-apis'
import type { Member } from '@/api/types'
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
  const apis = useInjectedApis()
  const [open, setOpen] = useState(false)
  const [members, setMembers] = useState<Member[]>([])

  useEffect(() => {
    if (!departmentId) {
      setMembers([])
      return
    }
    void apis.memberApi
      .list({ departmentId, page: 1, pageSize: 200 })
      .then((result) => setMembers(result.items))
      .catch(() => setMembers([]))
  }, [apis, departmentId])

  const selectedMembers = useMemo(
    () => members.filter((member) => selectedIds.includes(member.id)),
    [members, selectedIds],
  )

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
          disabled={!departmentId}
          aria-label="选择关联成员"
        >
          <Users className="size-4 shrink-0" />
          {selectedMembers.length === 0 ? (
            <span>{departmentId ? '选择成员…' : '请先选择团队'}</span>
          ) : (
            <span className="flex flex-wrap gap-1">
              {selectedMembers.map((member) => (
                <Badge
                  key={member.id}
                  variant="outline"
                  className="h-5 px-1.5 text-xs font-normal"
                >
                  {member.name}
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
                      aria-label={member.name}
                    />
                    <span className="flex-1 truncate">{member.name}</span>
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
