import { useEffect, useState } from 'react'
import type { ProjectView, Member } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { budgetApi } from '@/api/budget'
import { formatDisplayCurrency } from '@/lib/points'
import { BudgetMemberPicker } from './budget-member-picker'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { UserMinus, UserPlus } from 'lucide-react'

type ProjectMembersSectionProps = {
  project: ProjectView
  members: Member[]
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdateProject: (projectId: string, data: { memberIds: string[] }) => Promise<void>
  onUpdated: () => void
  getProjectMemberConsumed?: (projectId: string) => Promise<Record<string, number>>
}

export function ProjectMembersSection({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateProject,
  onUpdated,
  getProjectMemberConsumed,
}: ProjectMembersSectionProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draftMemberIds, setDraftMemberIds] = useState<string[]>([])
  const [savingMembers, setSavingMembers] = useState(false)
  const [consumedMap, setConsumedMap] = useState<Record<string, number>>({})

  useEffect(() => {
    let cancelled = false
    const fetchFn = getProjectMemberConsumed ?? budgetApi.getProjectMemberConsumed
    fetchFn(project.id)
      .then((data) => {
        if (!cancelled) setConsumedMap(data)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [project.id, getProjectMemberConsumed])

  function openDialog() {
    setDraftMemberIds([...project.memberIds])
    setDialogOpen(true)
  }

  async function saveMembers() {
    setSavingMembers(true)
    try {
      await onUpdateProject(project.id, { memberIds: draftMemberIds })
      setDialogOpen(false)
      onUpdated()
      toast.success('项目成员已更新')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '保存失败，请重试')
    } finally {
      setSavingMembers(false)
    }
  }

  async function removeMember(memberId: string) {
    const next = project.memberIds.filter((id) => id !== memberId)
    try {
      await onUpdateProject(project.id, { memberIds: next })
      onUpdated()
      toast.success('成员已移除')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '移除失败，请重试')
    }
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">关联成员</h4>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={openDialog}
          aria-label="编辑成员"
        >
          <UserPlus className="size-3.5" />
          编辑成员
        </Button>
      </div>

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
                const memberConsumed = consumedMap[m.id] ?? 0
                const memberPct =
                  project.consumed > 0 ? Math.round((memberConsumed / project.consumed) * 100) : 0
                return (
                  <TableRow key={m.id} className="even:bg-muted/40 hover:bg-muted/50">
                    <TableCell className="font-medium">{m.name}</TableCell>
                    <TableCell className="text-right tabular-nums">
                      {formatDisplayCurrency(memberConsumed)}
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
                        onClick={() => void removeMember(m.id)}
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

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>编辑项目成员</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <BudgetMemberPicker
              members={departmentMembers}
              loading={membersLoading}
              selectedIds={draftMemberIds}
              onChange={setDraftMemberIds}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setDialogOpen(false)}
              disabled={savingMembers}
            >
              取消
            </Button>
            <Button size="sm" onClick={saveMembers} disabled={savingMembers}>
              {savingMembers ? '保存中…' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
