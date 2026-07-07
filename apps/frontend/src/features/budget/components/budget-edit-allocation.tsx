import { useState } from 'react'
import type { BudgetNode, BudgetProjectView } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { nodeReservedPool } from '@/features/budget/lib/mappers'
import { cn } from '@/lib/utils'
import { Pencil, X, Check } from 'lucide-react'

interface BudgetEditAllocationProps {
  node: BudgetNode
  projects: BudgetProjectView[]
  overrunPolicyLabel: string
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

type RowDraft = {
  budget: string
}

export function BudgetEditAllocation({
  node,
  projects,
  overrunPolicyLabel,
  onUpdated,
  onUpdateDepartment,
}: BudgetEditAllocationProps) {
  const children = node.children ?? []
  const nodeProjects = projects.filter((project) => project.departmentId === node.id)

  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<Record<string, RowDraft>>({})
  const [reservedDraft, setReservedDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    const initial: Record<string, RowDraft> = {}
    for (const child of children) {
      initial[child.id] = { budget: String(child.budget) }
    }
    setReservedDraft(String(nodeReservedPool(node)))
    setDrafts(initial)
    setError(null)
    setEditing(true)
  }

  function cancelEdit() {
    setDrafts({})
    setError(null)
    setEditing(false)
  }

  function updateDraft(id: string, value: string) {
    setDrafts((prev) => ({ ...prev, [id]: { budget: value } }))
    setError(null)
  }

  function computeAllocated(draftBudgets: Record<string, number>, reservedPool: number) {
    const childSum = children.reduce(
      (sum, child) => sum + (draftBudgets[child.id] ?? child.budget),
      0,
    )
    const projectSum = nodeProjects.reduce((sum, project) => sum + project.budget, 0)
    return childSum + projectSum + reservedPool
  }

  function validate(): boolean {
    const draftBudgets: Record<string, number> = {}
    for (const child of children) {
      const raw = drafts[child.id]?.budget
      const value = raw !== undefined ? parseFloat(raw) : child.budget
      if (Number.isNaN(value) || value < 0) {
        setError(`"${child.name}" 额度无效`)
        return false
      }
      draftBudgets[child.id] = value
    }
    const reservedPool = parseFloat(reservedDraft)
    if (Number.isNaN(reservedPool) || reservedPool < 0) {
      setError('预留池余额无效')
      return false
    }
    const allocated = computeAllocated(draftBudgets, reservedPool)
    if (allocated > node.budget) {
      setError(
        `分配总额 ¥${allocated.toLocaleString()} 超出节点额度 ¥${node.budget.toLocaleString()}`,
      )
      return false
    }
    return true
  }

  async function handleSave() {
    if (!validate()) return
    setSaving(true)
    try {
      const reservedPool = parseFloat(reservedDraft)
      const updates: Promise<void>[] = []
      if (reservedPool !== nodeReservedPool(node)) {
        updates.push(onUpdateDepartment(node.id, { budget: node.budget, reservedPool }))
      }
      for (const child of children) {
        const nextBudget = parseFloat(drafts[child.id]?.budget ?? String(child.budget))
        if (nextBudget !== child.budget) {
          updates.push(onUpdateDepartment(child.id, { budget: nextBudget }))
        }
      }
      await Promise.all(updates)
      setEditing(false)
      setDrafts({})
      onUpdated()
    } catch {
      setError('保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  const draftBudgetMap: Record<string, number> = {}
  for (const child of children) {
    const raw = drafts[child.id]?.budget
    const value = raw !== undefined ? parseFloat(raw) : Number.NaN
    draftBudgetMap[child.id] = Number.isNaN(value) ? child.budget : value
  }
  const reservedValue = editing ? parseFloat(reservedDraft) : nodeReservedPool(node)
  const projectSum = nodeProjects.reduce((sum, project) => sum + project.budget, 0)
  const usedSum =
    Object.values(draftBudgetMap).reduce((sum, value) => sum + value, 0) +
    projectSum +
    (Number.isNaN(reservedValue) ? 0 : reservedValue)
  const remaining = node.budget - usedSum

  if (children.length === 0 && nodeProjects.length === 0) return null

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">子节点分配</h4>
        {!editing ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1.5 text-xs text-muted-foreground"
            onClick={startEdit}
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
              onClick={cancelEdit}
              disabled={saving}
              aria-label="取消编辑"
            >
              <X className="size-3.5" />
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={handleSave}
              disabled={saving}
              aria-label="保存分配"
            >
              <Check className="size-3.5" />
              保存
            </Button>
          </div>
        )}
      </div>

      <div className="rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow className="border-border/50 hover:bg-transparent">
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                名称
              </TableHead>
              <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                额度
              </TableHead>
              <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
                已消耗
              </TableHead>
              <TableHead className="w-32 text-xs font-medium uppercase text-muted-foreground">
                进度
              </TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">
                超限策略
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {children.map((child) => {
              const childPct =
                child.budget > 0 ? Math.round((child.consumed / child.budget) * 100) : 0
              const draftBudget = drafts[child.id]?.budget ?? String(child.budget)
              const draftValue = parseFloat(draftBudget)
              const budgetOver = editing && !Number.isNaN(draftValue) && draftValue > node.budget

              return (
                <TableRow key={child.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium">{child.name}</TableCell>
                  <TableCell className="text-right">
                    {editing ? (
                      <Input
                        type="number"
                        min={0}
                        value={draftBudget}
                        onChange={(event) => updateDraft(child.id, event.target.value)}
                        className={cn(
                          'h-7 w-28 text-right tabular-nums',
                          budgetOver && 'border-red-500 focus-visible:ring-red-500/30',
                        )}
                      />
                    ) : (
                      <span className="tabular-nums">¥{child.budget.toLocaleString()}</span>
                    )}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    ¥{child.consumed.toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={childPct} className="h-1.5 flex-1" />
                      <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                        {childPct}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs">
                      {overrunPolicyLabel}
                    </Badge>
                  </TableCell>
                </TableRow>
              )
            })}
            {nodeProjects.map((project) => {
              const projectPct =
                project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0
              return (
                <TableRow key={project.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium text-muted-foreground">
                    {project.name}
                    <span className="ml-1.5 text-xs text-muted-foreground/60">(项目)</span>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    ¥{project.budget.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    ¥{project.consumed.toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={projectPct} className="h-1.5 flex-1" />
                      <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                        {projectPct}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs">
                      {overrunPolicyLabel}
                    </Badge>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>

        {editing && (
          <div className="border-t border-border px-4 py-3">
            <label className="mb-1 block text-xs text-muted-foreground">预留池余额</label>
            <Input
              type="number"
              min={0}
              value={reservedDraft}
              onChange={(event) => {
                setReservedDraft(event.target.value)
                setError(null)
              }}
              className="h-7 w-40 tabular-nums"
            />
          </div>
        )}

        <div className="flex items-center justify-between border-t border-border px-4 py-2.5">
          {error ? <p className="text-xs text-red-600">{error}</p> : <span />}
          <p
            className={cn(
              'text-xs tabular-nums',
              remaining < 0 ? 'text-red-600' : 'text-muted-foreground',
            )}
          >
            剩余可分配：¥{remaining.toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  )
}
