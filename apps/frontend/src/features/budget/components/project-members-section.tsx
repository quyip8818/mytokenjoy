import { useEffect, useMemo, useState } from 'react'
import type { Member, ProjectView } from '@/api/types'
import { budgetApi } from '@/api/budget'
import { formatDisplayCurrency } from '@/lib/points'
import { BudgetMemberPicker } from './budget-member-picker'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Check, KeyRound, UserMinus, UserPlus, X } from 'lucide-react'

type ProjectMembersSectionProps = {
  project: ProjectView
  members: Member[]
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdateProject: (
    projectId: string,
    data: { memberIds?: string[]; memberBudgets?: Record<string, number> },
  ) => Promise<void>
  onCreateMemberKey?: (member: Member) => void
  onUpdated: () => void
  getProjectMemberConsumed?: (projectId: string) => Promise<Record<string, number>>
}

export function ProjectMembersSection({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateProject,
  onCreateMemberKey,
  onUpdated,
  getProjectMemberConsumed,
}: ProjectMembersSectionProps) {
  const [editingMembers, setEditingMembers] = useState(false)
  const [draftMemberIds, setDraftMemberIds] = useState<string[]>([])
  const [savingMembers, setSavingMembers] = useState(false)
  const [membersError, setMembersError] = useState<string | null>(null)
  const [consumedMap, setConsumedMap] = useState<Record<string, number>>({})
  const [draftBudgetOverrides, setDraftBudgetOverrides] = useState<Record<string, string>>({})
  const [savingBudgetFor, setSavingBudgetFor] = useState<string | null>(null)

  const baseDraftBudgets = useMemo(() => {
    const next: Record<string, string> = {}
    for (const member of members) {
      const budget = project.memberBudgets?.[member.id] ?? 0
      next[member.id] = String(budget)
    }
    return next
  }, [members, project.memberBudgets])

  const draftBudgetFor = (memberId: string) =>
    draftBudgetOverrides[memberId] ?? baseDraftBudgets[memberId] ?? '0'

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
      await onUpdateProject(project.id, { memberIds: draftMemberIds })
      setEditingMembers(false)
      onUpdated()
    } catch {
      setMembersError('保存失败，请重试')
    } finally {
      setSavingMembers(false)
    }
  }

  async function saveMemberBudget(memberId: string) {
    const value = Number(draftBudgetFor(memberId))
    if (Number.isNaN(value) || value < 0) return
    setSavingBudgetFor(memberId)
    try {
      await onUpdateProject(project.id, { memberBudgets: { [memberId]: value } })
      setDraftBudgetOverrides((prev) => {
        const next = { ...prev }
        delete next[memberId]
        return next
      })
      onUpdated()
    } finally {
      setSavingBudgetFor(null)
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
                  子额度
                </TableHead>
                <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                  已消耗
                </TableHead>
                <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                  剩余
                </TableHead>
                <TableHead className="w-28 text-right text-xs font-medium uppercase text-muted-foreground">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-6 text-center text-xs text-muted-foreground">
                    暂无关联成员
                  </TableCell>
                </TableRow>
              ) : (
                members.map((m) => {
                  const memberBudget = project.memberBudgets?.[m.id] ?? 0
                  const memberConsumed = consumedMap[m.id] ?? 0
                  const memberRemaining = Math.max(0, memberBudget - memberConsumed)
                  return (
                    <TableRow key={m.id} className="even:bg-muted/40 hover:bg-muted/50">
                      <TableCell className="font-medium">{m.name}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Input
                            type="number"
                            className="h-7 w-24 text-right text-xs"
                            value={draftBudgetFor(m.id)}
                            onChange={(e) =>
                              setDraftBudgetOverrides((prev) => ({ ...prev, [m.id]: e.target.value }))
                            }
                          />
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7"
                            disabled={savingBudgetFor === m.id}
                            aria-label={`保存 ${m.name} 子额度`}
                            onClick={() => void saveMemberBudget(m.id)}
                          >
                            <Check className="size-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {formatDisplayCurrency(memberConsumed)}
                      </TableCell>
                      <TableCell className="text-right tabular-nums text-muted-foreground">
                        {formatDisplayCurrency(memberRemaining)}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7 text-muted-foreground"
                            disabled={memberBudget <= 0 || !onCreateMemberKey}
                            aria-label={`为 ${m.name} 签发项目成员 Key`}
                            onClick={() => onCreateMemberKey?.(m)}
                          >
                            <KeyRound className="size-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7 text-muted-foreground hover:text-red-600"
                            aria-label={`移除成员 ${m.name}`}
                            onClick={async () => {
                              const next = project.memberIds.filter((id) => id !== m.id)
                              try {
                                await onUpdateProject(project.id, { memberIds: next })
                                onUpdated()
                              } catch {
                                // silent
                              }
                            }}
                          >
                            <UserMinus className="size-3.5" />
                          </Button>
                        </div>
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
