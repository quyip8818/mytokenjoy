import { useMemo, useState } from 'react'
import type { BudgetNode, BudgetProjectView } from '@/api/types'
import { nodeReservedPool } from '../lib/mappers'

type RowDraft = {
  budget: string
}

type UseBudgetAllocationEditOptions = {
  node: BudgetNode
  projects: BudgetProjectView[]
  onUpdated: () => void
  onUpdateDepartment: (
    departmentId: string,
    data: { budget: number; reservedPool?: number },
  ) => Promise<void>
}

export function useBudgetAllocationEdit({
  node,
  projects,
  onUpdated,
  onUpdateDepartment,
}: UseBudgetAllocationEditOptions) {
  const children = useMemo(() => node.children ?? [], [node.children])
  const nodeProjects = useMemo(
    () => projects.filter((project) => project.departmentId === node.id),
    [projects, node.id],
  )

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

  function updateReservedDraft(value: string) {
    setReservedDraft(value)
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

  return {
    children,
    nodeProjects,
    editing,
    drafts,
    reservedDraft,
    saving,
    error,
    remaining,
    startEdit,
    cancelEdit,
    updateDraft,
    updateReservedDraft,
    handleSave,
  }
}
