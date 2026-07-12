import { useState } from 'react'
import type { ProjectView, Member } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { BudgetMemberPicker } from './budget-member-picker'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
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
}

export function ProjectMembersSection({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateProject,
  onUpdated,
}: ProjectMembersSectionProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draftMemberIds, setDraftMemberIds] = useState<string[]>([])
  const [savingMembers, setSavingMembers] = useState(false)
  const [removeTarget, setRemoveTarget] = useState<Member | null>(null)

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

  async function confirmRemove() {
    if (!removeTarget) return
    const next = project.memberIds.filter((id) => id !== removeTarget.id)
    try {
      await onUpdateProject(project.id, { memberIds: next })
      onUpdated()
      toast.success(`已移除成员「${removeTarget.name}」`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '移除失败，请重试')
    } finally {
      setRemoveTarget(null)
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
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                部门
              </TableHead>
              <TableHead className="w-10" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} className="py-6 text-center text-xs text-muted-foreground">
                  暂无关联成员
                </TableCell>
              </TableRow>
            ) : (
              members.map((m) => (
                <TableRow key={m.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium">{m.name}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{m.departmentName}</TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6 text-muted-foreground hover:text-red-600"
                      aria-label={`移除成员 ${m.name}`}
                      onClick={() => setRemoveTarget(m)}
                    >
                      <UserMinus className="size-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
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
            <Button variant="outline" size="sm" onClick={() => setDialogOpen(false)} disabled={savingMembers}>
              取消
            </Button>
            <Button size="sm" onClick={saveMembers} disabled={savingMembers}>
              {savingMembers ? '保存中…' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!removeTarget} onOpenChange={(open) => { if (!open) setRemoveTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认移除成员</AlertDialogTitle>
            <AlertDialogDescription>
              确定将「{removeTarget?.name}」从项目中移除吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setRemoveTarget(null)}>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => void confirmRemove()}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              移除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
