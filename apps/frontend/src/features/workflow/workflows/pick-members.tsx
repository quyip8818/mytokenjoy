import { useEffect, useState } from 'react'
import type { Member } from '@/api/types'
import { memberApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { WORKFLOW_LIST_ITEM_SELECTED_CLASS } from '../constants'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'

export function PickMembersWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'pick-members'>) {
  const departmentId = entry.payload.departmentId as string
  const selectedIds = (entry.payload.selectedIds as string[]) ?? []
  const onConfirm = entry.payload.onConfirm as
    | ((memberIds: string[], members: Member[]) => void)
    | undefined

  const [members, setMembers] = useState<Member[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set(selectedIds))

  useEffect(() => {
    if (!departmentId) return
    let cancelled = false
    void memberApi.list({ departmentId, page: 1, pageSize: 100 }).then((res) => {
      if (!cancelled) setMembers(res.items)
    })
    return () => {
      cancelled = true
    }
  }, [departmentId])

  const toggleMember = (member: Member) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(member.id)) next.delete(member.id)
      else next.add(member.id)
      return next
    })
    onSetDirty(true)
  }

  const handleConfirm = () => {
    const picked = members.filter((m) => selected.has(m.id))
    if (picked.length === 0) return
    onConfirm?.(
      picked.map((m) => m.id),
      picked,
    )
    onPop()
  }

  return (
    <WorkflowPanelChrome
      title="选择成员"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel="确认"
          onPrimary={handleConfirm}
          primaryDisabled={selected.size === 0}
        />
      }
    >
      <WorkflowFormLayout variant="full">
        {!departmentId && (
          <p className="text-sm text-amber-700 bg-amber-50 border border-amber-200 rounded-md px-3 py-2">
            请先选择来源部门
          </p>
        )}
        <div className="max-h-[50vh] overflow-y-auto rounded-lg border border-border/60 divide-y divide-border/40">
          {members.length === 0 ? (
            <p className="px-4 py-8 text-center text-sm text-muted-foreground">
              {departmentId ? '该部门暂无成员' : '请先选择部门'}
            </p>
          ) : (
            members.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => toggleMember(m)}
                className={cn(
                  'flex w-full items-center gap-3 px-4 py-3 text-left hover:bg-primary/5',
                  selected.has(m.id) && WORKFLOW_LIST_ITEM_SELECTED_CLASS,
                )}
              >
                <Checkbox checked={selected.has(m.id)} onCheckedChange={() => toggleMember(m)} />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">{m.name}</p>
                  <p className="text-xs text-muted-foreground truncate">{m.departmentName}</p>
                </div>
              </button>
            ))
          )}
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
