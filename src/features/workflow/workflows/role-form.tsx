import { useState } from 'react'
import { toast } from 'sonner'
import type { Permission, Role } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '../use-workflow'

function RoleFormInner({
  role,
  permissions,
  onSubmit,
  onClose,
  onSetDirty,
  onPush,
}: {
  role: Role | null | undefined
  permissions: Permission[]
  onSubmit?: (data: { name: string; permissions: string[] }) => Promise<void>
  onClose: () => void
  onSetDirty: (dirty: boolean) => void
  onPush: WorkflowComponentProps['onPush']
}) {
  const { closeAll } = useWorkflow()
  const [name, setName] = useState(role?.name ?? '')
  const [selectedPerms, setSelectedPerms] = useState<string[]>(role?.permissions ?? [])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const openPermissionPicker = () => {
    onPush('permission-picker', {
      permissions,
      selected: selectedPerms,
      onConfirm: (perms: string[]) => {
        setSelectedPerms(perms)
        onSetDirty(true)
      },
    })
  }

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('请输入角色名称')
      return
    }
    setSubmitting(true)
    try {
      await onSubmit?.({ name: name.trim(), permissions: selectedPerms })
      toast.success(role ? '角色已更新' : '角色已创建')
      closeAll()
    } catch {
      toast.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title={role ? '编辑角色' : '创建角色'}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={role ? '保存' : '创建'}
          onPrimary={handleSubmit}
          primaryDisabled={submitting}
        />
      }
    >
      <div className="max-w-md space-y-4">
        <div className="space-y-1.5">
          <Label>角色名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              setError('')
              onSetDirty(true)
            }}
            placeholder="输入角色名称"
          />
          {error && <p className="text-xs text-destructive">{error}</p>}
        </div>
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label>权限配置</Label>
            <Button type="button" variant="outline" size="sm" onClick={openPermissionPicker}>
              配置权限
            </Button>
          </div>
          <p className="text-sm text-muted-foreground">已选 {selectedPerms.length} 项权限</p>
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}

export function RoleFormWorkflow({ entry, onClose, onSetDirty, onPush }: WorkflowComponentProps) {
  const role = entry.payload.role as Role | null | undefined
  const permissions = (entry.payload.permissions as Permission[]) ?? []
  const onSubmit = entry.payload.onSubmit as
    | ((data: { name: string; permissions: string[] }) => Promise<void>)
    | undefined
  const formKey = role?.id ?? 'new'

  return (
    <RoleFormInner
      key={formKey}
      role={role}
      permissions={permissions}
      onSubmit={onSubmit}
      onClose={onClose}
      onSetDirty={onSetDirty}
      onPush={onPush}
    />
  )
}
