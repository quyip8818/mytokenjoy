import { useState, useMemo } from 'react'
import { budgetApi } from '@/api/budget'
import type { BudgetProject, OverrunPolicy } from '@/api/types'
import { mockMembers } from '@/mocks/data'
import { BudgetMemberPicker } from './budget-member-picker'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { Check, Pencil, Trash2, UserMinus, UserPlus, X } from 'lucide-react'

interface BudgetDetailProjectProps {
  project: BudgetProject
  onUpdated: () => void
  onDeleted: () => void
}

const POLICY_LABELS: Record<OverrunPolicy, { label: string; className: string }> = {
  hard_reject: { label: '硬拒绝', className: 'bg-red-50 text-red-700 border-red-200' },
  approval: { label: '审批追加', className: 'bg-primary/10 text-primary border-primary/20' },
  downgrade: { label: '降级路由', className: 'bg-amber-50 text-amber-700 border-amber-200' },
}

function SummaryCard({
  label,
  value,
  muted,
  highlight,
}: {
  label: string
  value: number
  muted?: boolean
  highlight?: boolean
}) {
  return (
    <div className="rounded-lg border border-border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p
        className={cn(
          'mt-1 text-lg font-semibold tabular-nums',
          highlight
            ? 'text-red-600'
            : muted
              ? 'text-muted-foreground'
              : 'text-foreground'
        )}
      >
        ¥{value.toLocaleString()}
      </p>
    </div>
  )
}

export function BudgetDetailProject({
  project,
  onUpdated,
  onDeleted,
}: BudgetDetailProjectProps) {
  const policy = POLICY_LABELS[project.overrunPolicy]
  const remaining = project.budget - project.consumed
  const pct = project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0

  // Settings inline edit state
  const [editingSettings, setEditingSettings] = useState(false)
  const [draftBudget, setDraftBudget] = useState('')
  const [draftPolicy, setDraftPolicy] = useState<OverrunPolicy>(project.overrunPolicy)
  const [settingsError, setSettingsError] = useState<string | null>(null)
  const [savingSettings, setSavingSettings] = useState(false)

  // Member management state
  const [editingMembers, setEditingMembers] = useState(false)
  const [draftMemberIds, setDraftMemberIds] = useState<string[]>([])
  const [savingMembers, setSavingMembers] = useState(false)
  const [membersError, setMembersError] = useState<string | null>(null)

  // Deleting state
  const [deleting, setDeleting] = useState(false)

  const members = useMemo(
    () => mockMembers.filter((m) => project.memberIds.includes(m.id)),
    [project.memberIds]
  )

  function startEditSettings() {
    setDraftBudget(String(project.budget))
    setDraftPolicy(project.overrunPolicy)
    setSettingsError(null)
    setEditingSettings(true)
  }

  function cancelEditSettings() {
    setEditingSettings(false)
    setSettingsError(null)
  }

  async function saveSettings() {
    setSettingsError(null)
    const budgetNum = parseFloat(draftBudget)
    if (!draftBudget || isNaN(budgetNum) || budgetNum < 0) {
      setSettingsError('请输入有效的额度')
      return
    }
    if (budgetNum < project.consumed) {
      setSettingsError(`额度不能低于已消耗 ¥${project.consumed.toLocaleString()}`)
      return
    }
    setSavingSettings(true)
    try {
      await budgetApi.updateProject(project.id, {
        budget: budgetNum,
        overrunPolicy: draftPolicy,
      })
      setEditingSettings(false)
      onUpdated()
    } catch {
      setSettingsError('保存失败，请重试')
    } finally {
      setSavingSettings(false)
    }
  }

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
      await budgetApi.updateProject(project.id, { memberIds: draftMemberIds })
      setEditingMembers(false)
      onUpdated()
    } catch {
      setMembersError('保存失败，请重试')
    } finally {
      setSavingMembers(false)
    }
  }

  async function handleDelete() {
    setDeleting(true)
    try {
      await budgetApi.deleteProject(project.id)
      onDeleted()
    } catch {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      {/* Header */}
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-semibold text-foreground">{project.name}</h3>
          <Badge variant="outline" className="text-xs font-normal">
            所属：{project.departmentName}
          </Badge>
          <Badge variant="outline" className={cn(policy.className, 'text-xs')}>
            {policy.label}
          </Badge>
        </div>

        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 text-muted-foreground hover:text-red-600"
              aria-label="删除项目"
              disabled={deleting}
            >
              <Trash2 className="size-4" />
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogMedia>
                <Trash2 />
              </AlertDialogMedia>
              <AlertDialogTitle>删除项目</AlertDialogTitle>
              <AlertDialogDescription>
                确定要删除项目「{project.name}」吗？此操作不可撤销。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction
                variant="destructive"
                onClick={handleDelete}
                disabled={deleting}
              >
                {deleting ? '删除中…' : '删除'}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-3 gap-4">
        <SummaryCard label="项目额度" value={project.budget} />
        <SummaryCard label="已消耗" value={project.consumed} muted />
        <SummaryCard label="剩余" value={remaining} highlight={remaining < 0} />
      </div>

      {/* Members section */}
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
              departmentId={project.departmentId}
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
                    <TableCell
                      colSpan={4}
                      className="py-6 text-center text-xs text-muted-foreground"
                    >
                      暂无关联成员
                    </TableCell>
                  </TableRow>
                ) : (
                  members.map((m) => {
                    const memberConsumed = 0
                    const memberPct =
                      project.consumed > 0
                        ? Math.round((memberConsumed / project.consumed) * 100)
                        : 0
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
                                await budgetApi.updateProject(project.id, { memberIds: next })
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

      {/* Settings section */}
      <div>
        <div className="mb-3 flex items-center justify-between">
          <h4 className="text-sm font-semibold text-foreground">项目设置</h4>
          {!editingSettings ? (
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs text-muted-foreground"
              onClick={startEditSettings}
              aria-label="编辑项目设置"
            >
              <Pencil className="size-3.5" />
              编辑
            </Button>
          ) : (
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 gap-1.5 text-xs text-muted-foreground"
                onClick={cancelEditSettings}
                disabled={savingSettings}
                aria-label="取消编辑设置"
              >
                <X className="size-3.5" />
                取消
              </Button>
              <Button
                size="sm"
                className="h-7 gap-1.5 text-xs"
                onClick={saveSettings}
                disabled={savingSettings}
                aria-label="保存设置"
              >
                <Check className="size-3.5" />
                保存
              </Button>
            </div>
          )}
        </div>

        <div className="rounded-lg border border-border p-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-1.5">
              <Label htmlFor="proj-settings-budget" className="text-xs text-muted-foreground">
                项目额度（元）
              </Label>
              {editingSettings ? (
                <Input
                  id="proj-settings-budget"
                  type="number"
                  min={0}
                  value={draftBudget}
                  onChange={(e) => setDraftBudget(e.target.value)}
                  className="h-8 text-sm tabular-nums"
                />
              ) : (
                <p className="text-sm font-medium tabular-nums">
                  ¥{project.budget.toLocaleString()}
                </p>
              )}
            </div>

            <div className="grid gap-1.5">
              <Label htmlFor="proj-settings-policy" className="text-xs text-muted-foreground">
                超限策略
              </Label>
              {editingSettings ? (
                <Select
                  value={draftPolicy}
                  onValueChange={(v) => setDraftPolicy(v as OverrunPolicy)}
                >
                  <SelectTrigger id="proj-settings-policy" size="sm" className="h-8 w-full text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hard_reject">硬拒绝</SelectItem>
                    <SelectItem value="approval">审批追加</SelectItem>
                    <SelectItem value="downgrade">降级路由</SelectItem>
                  </SelectContent>
                </Select>
              ) : (
                <Badge
                  variant="outline"
                  className={cn(policy.className, 'w-fit text-xs font-normal')}
                >
                  {policy.label}
                </Badge>
              )}
            </div>
          </div>

          {editingSettings && settingsError && (
            <p className="mt-2 text-xs text-red-600">{settingsError}</p>
          )}

          {!editingSettings && (
            <div className="mt-3 border-t border-border pt-3">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>本月消耗进度</span>
                <span className="tabular-nums">
                  ¥{project.consumed.toLocaleString()} / ¥{project.budget.toLocaleString()} ({pct}%)
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
