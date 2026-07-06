import { useState } from 'react'
import { budgetApi } from '@/api/budget'
import type { BudgetNode, BudgetProject, OverrunPolicy } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { Pencil, X, Check } from 'lucide-react'

interface BudgetEditAllocationProps {
  node: BudgetNode
  projects: BudgetProject[]
  onUpdated: () => void
}

const POLICY_LABELS: Record<OverrunPolicy, { label: string; className: string }> = {
  hard_reject: { label: '硬拒绝', className: 'bg-red-50 text-red-700 border-red-200' },
  approval: { label: '审批追加', className: 'bg-primary/10 text-primary border-primary/20' },
  downgrade: { label: '降级路由', className: 'bg-amber-50 text-amber-700 border-amber-200' },
}

type RowDraft = {
  budget: string
  overrunPolicy: OverrunPolicy
}

export function BudgetEditAllocation({ node, projects, onUpdated }: BudgetEditAllocationProps) {
  const children = node.children ?? []
  const nodeProjects = projects.filter((p) => p.departmentId === node.id)

  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<Record<string, RowDraft>>({})
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    const initial: Record<string, RowDraft> = {}
    for (const c of children) {
      initial[c.id] = { budget: String(c.budget), overrunPolicy: c.overrunPolicy }
    }
    setDrafts(initial)
    setError(null)
    setEditing(true)
  }

  function cancelEdit() {
    setDrafts({})
    setError(null)
    setEditing(false)
  }

  function updateDraft(id: string, field: keyof RowDraft, value: string) {
    setDrafts((prev) => ({ ...prev, [id]: { ...prev[id], [field]: value } }))
    setError(null)
  }

  function computeAllocated(draftBudgets: Record<string, number>) {
    const childSum = children.reduce((s, c) => s + (draftBudgets[c.id] ?? c.budget), 0)
    const projectSum = nodeProjects.reduce((s, p) => s + p.budget, 0)
    return childSum + projectSum + node.reserved
  }

  function validate(): boolean {
    const draftBudgets: Record<string, number> = {}
    for (const c of children) {
      const raw = drafts[c.id]?.budget
      const val = raw !== undefined ? parseFloat(raw) : c.budget
      if (isNaN(val) || val < 0) {
        setError(`"${c.name}" 额度无效`)
        return false
      }
      draftBudgets[c.id] = val
    }
    const allocated = computeAllocated(draftBudgets)
    if (allocated > node.budget) {
      setError(`分配总额 ¥${allocated.toLocaleString()} 超出节点额度 ¥${node.budget.toLocaleString()}`)
      return false
    }
    return true
  }

  async function handleSave() {
    if (!validate()) return
    setSaving(true)
    try {
      const changed = children.filter((c) => {
        const d = drafts[c.id]
        if (!d) return false
        return parseFloat(d.budget) !== c.budget || d.overrunPolicy !== c.overrunPolicy
      })
      await Promise.all(
        changed.map((c) =>
          budgetApi.updateNode(c.id, {
            budget: parseFloat(drafts[c.id].budget),
            overrunPolicy: drafts[c.id].overrunPolicy,
          })
        )
      )
      setEditing(false)
      setDrafts({})
      onUpdated()
    } catch {
      setError('保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  // Compute remaining for footer
  const draftBudgetMap: Record<string, number> = {}
  for (const c of children) {
    const raw = drafts[c.id]?.budget
    const val = raw !== undefined ? parseFloat(raw) : NaN
    draftBudgetMap[c.id] = isNaN(val) ? c.budget : val
  }
  const projectSum = nodeProjects.reduce((s, p) => s + p.budget, 0)
  const usedSum = Object.values(draftBudgetMap).reduce((s, v) => s + v, 0) + projectSum + node.reserved
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
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">名称</TableHead>
              <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">额度</TableHead>
              <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">已消耗</TableHead>
              <TableHead className="w-32 text-xs font-medium uppercase text-muted-foreground">进度</TableHead>
              <TableHead className="text-xs font-medium uppercase text-muted-foreground">超限策略</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {children.map((child) => {
              const pct = child.budget > 0 ? Math.round((child.consumed / child.budget) * 100) : 0
              const policy = POLICY_LABELS[child.overrunPolicy]
              const draft = drafts[child.id]
              const draftBudget = draft?.budget ?? String(child.budget)
              const draftPolicy = draft?.overrunPolicy ?? child.overrunPolicy
              const draftVal = parseFloat(draftBudget)
              const budgetOver = editing && !isNaN(draftVal) && draftVal > node.budget

              return (
                <TableRow key={child.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium">{child.name}</TableCell>
                  <TableCell className="text-right">
                    {editing ? (
                      <Input
                        type="number"
                        min={0}
                        value={draftBudget}
                        onChange={(e) => updateDraft(child.id, 'budget', e.target.value)}
                        className={cn(
                          'h-7 w-28 text-right tabular-nums',
                          budgetOver && 'border-red-500 focus-visible:ring-red-500/30'
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
                      <Progress value={pct} className="h-1.5 flex-1" />
                      <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                        {pct}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    {editing ? (
                      <Select
                        value={draftPolicy}
                        onValueChange={(v) =>
                          updateDraft(child.id, 'overrunPolicy', v)
                        }
                      >
                        <SelectTrigger size="sm" className="h-7 w-28 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="hard_reject">硬拒绝</SelectItem>
                          <SelectItem value="approval">审批追加</SelectItem>
                          <SelectItem value="downgrade">降级路由</SelectItem>
                        </SelectContent>
                      </Select>
                    ) : (
                      <Badge variant="outline" className={cn(policy.className, 'text-xs')}>
                        {policy.label}
                      </Badge>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
            {nodeProjects.map((proj) => {
              const pct = proj.budget > 0 ? Math.round((proj.consumed / proj.budget) * 100) : 0
              const policy = POLICY_LABELS[proj.overrunPolicy]
              return (
                <TableRow key={proj.id} className="even:bg-muted/40 hover:bg-muted/50">
                  <TableCell className="font-medium text-muted-foreground">
                    {proj.name}
                    <span className="ml-1.5 text-xs text-muted-foreground/60">(项目)</span>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    ¥{proj.budget.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    ¥{proj.consumed.toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress value={pct} className="h-1.5 flex-1" />
                      <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                        {pct}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={cn(policy.className, 'text-xs')}>
                      {policy.label}
                    </Badge>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>

        <div className="flex items-center justify-between border-t border-border px-4 py-2.5">
          {error ? (
            <p className="text-xs text-red-600">{error}</p>
          ) : (
            <span />
          )}
          <p className={cn('text-xs tabular-nums', remaining < 0 ? 'text-red-600' : 'text-muted-foreground')}>
            剩余可分配：¥{remaining.toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  )
}
