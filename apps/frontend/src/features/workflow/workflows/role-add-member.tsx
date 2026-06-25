import { useState } from 'react'
import { toast } from 'sonner'
import { X } from 'lucide-react'
import type { Member } from '@/api/types'
import { roleApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useWorkflow } from '../use-workflow'

export function RoleAddMemberWorkflow({
  entry,
  onClose,
  onSetDirty,
  onPush,
}: WorkflowComponentProps<'role-add-member'>) {
  const { closeAll } = useWorkflow()
  const roleId = entry.payload.roleId as string
  const roleName = (entry.payload.roleName as string) ?? ''
  const existingMemberIds = (entry.payload.existingMemberIds as string[]) ?? []
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined

  const [pending, setPending] = useState<Member[]>([])
  const [submitting, setSubmitting] = useState(false)

  const excludeIds = [...existingMemberIds, ...pending.map((m) => m.id)]

  const openSearch = () => {
    onPush('member-search', {
      excludeIds,
      multi: true,
      onConfirm: (members: Member[]) => {
        setPending((prev) => {
          const ids = new Set(prev.map((m) => m.id))
          const added = members.filter((m) => !ids.has(m.id))
          return [...prev, ...added]
        })
        onSetDirty(true)
      },
    })
  }

  const removePending = (id: string) => {
    setPending((prev) => prev.filter((m) => m.id !== id))
    onSetDirty(true)
  }

  const handleSubmit = async () => {
    if (pending.length === 0) return
    setSubmitting(true)
    try {
      for (const m of pending) {
        await roleApi.addMember(roleId, m.id)
      }
      toast.success(`已添加 ${pending.length} 名成员`)
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('添加失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="添加角色成员"
      onClose={onClose}
      contextBar={roleName ? `角色：${roleName}` : undefined}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel="确认添加"
          onPrimary={handleSubmit}
          primaryDisabled={pending.length === 0 || submitting}
        />
      }
    >
      <WorkflowFormLayout variant="full">
        <Button variant="outline" size="sm" onClick={openSearch}>
          搜索并添加成员
        </Button>
        {pending.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4">尚未选择成员</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {pending.map((m) => (
              <Badge key={m.id} variant="secondary" className="gap-1 pr-1">
                {m.name}
                <button
                  type="button"
                  onClick={() => removePending(m.id)}
                  className="rounded-full hover:bg-muted p-0.5"
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            ))}
          </div>
        )}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
