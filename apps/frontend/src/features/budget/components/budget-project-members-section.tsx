import { useState } from 'react'
import type { BudgetProjectView, Member } from '@/api/types'
import { BudgetMemberPicker } from './budget-member-picker'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Check, UserMinus, UserPlus, X } from 'lucide-react'

type BudgetProjectMembersSectionProps = {
  project: BudgetProjectView
  members: Member[]
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdateGroup: (groupId: string, data: { memberIds: string[] }) => Promise<void>
  onUpdated: () => void
}

export function BudgetProjectMembersSection({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateGroup,
  onUpdated,
}: BudgetProjectMembersSectionProps) {
  const [editingMembers, setEditingMembers] = useState(false)
  const [draftMemberIds, setDraftMemberIds] = useState<string[]>([])
  const [savingMembers, setSavingMembers] = useState(false)
  const [membersError, setMembersError] = useState<string | null>(null)

  function startEditMembers() {
    setDraftMemberIds([...project.memberIds])
    setMembersError(null)
    setEditingMembers(true)
  }

  function cancelEditMembers() {
    setEditingMembers(false)
    setMembersError(null)
  }

  async function saveMembers() {
    setSavingMembers(true)
    setMembersError(null)
    try {
      await onUpdateGroup(project.id, { memberIds: draftMemberIds })
      setEditingMembers(false)
      onUpdated()
    } catch {
      setMembersError('保存失败，请重试')
    } finally {
      setSavingMembers(false)
    }
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">关联成员</h4>
        {!editingMembers ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 text-xs text-muted-foreground"
            onClick={startEditMembers}
            aria-label="编辑成员"
          >
            <UserPlus className="size-3.5" />
            编辑成员
          </Button>
        ) : (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs text-muted-foreground"
              onClick={cancelEditMembers}
              disabled={savingMembers}
              aria-label="取消编辑成员"
            >
              <X className="size-3.5" />
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={saveMembers}
              disabled={savingMembers}
              aria-label="保存成员"
            >
              <Check className="size-3.5" />
              保存
            </Button>
          </div>
        )}
      </div>

      {editingMembers ? (
        <div className="space-y-2">
          <BudgetMemberPicker
            members={departmentMembers}
            loading={membersLoading}
            selectedIds={draftMemberIds}
            onChange={setDraftMemberIds}
          />
          {membersError && <p className="text-xs text-red-600">{membersError}</p>}
        </div>
      ) : (
        <div className="rounded-lg border border-border">
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                  成员
                </TableHead>
                <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                  已消耗
                </TableHead>
                <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                  占比
                </TableHead>
                <TableHead className="w-10" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="py-6 text-center text-xs text-muted-foreground">
                    暂无关联成员
                  </TableCell>
                </TableRow>
              ) : (
                members.map((m) => {
                  const memberConsumed = 0
                  const memberPct =
                    project.consumed > 0 ? Math.round((memberConsumed / project.consumed) * 100) : 0
                  return (
                    <TableRow key={m.id} className="even:bg-muted/40 hover:bg-muted/50">
                      <TableCell className="font-medium">{m.name}</TableCell>
                      <TableCell className="text-right tabular-nums">
                        ¥{memberConsumed.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right tabular-nums text-muted-foreground">
                        {memberPct}%
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6 text-muted-foreground hover:text-red-600"
                          aria-label={`移除成员 ${m.name}`}
                          onClick={async () => {
                            const next = project.memberIds.filter((id) => id !== m.id)
                            try {
                              await onUpdateGroup(project.id, { memberIds: next })
                              onUpdated()
                            } catch {
                              // silent — optimistic UX not required here
                            }
                          }}
                        >
                          <UserMinus className="size-3.5" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
