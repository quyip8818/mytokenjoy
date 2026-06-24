import { useState } from 'react'
import { toast } from 'sonner'
import type { Department } from '@/api/types'
import { departmentApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useWorkflow } from '../use-workflow'

function DeptFormInner({
  department,
  parentId,
  parentName,
  onClose,
  onSetDirty,
  onSuccess,
}: {
  department: Department | null | undefined
  parentId: string
  parentName: string
  onClose: () => void
  onSetDirty: (dirty: boolean) => void
  onSuccess?: () => void
}) {
  const { closeAll } = useWorkflow()
  const [name, setName] = useState(department?.name ?? '')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('请输入部门名称')
      return
    }
    setSubmitting(true)
    try {
      if (department) {
        await departmentApi.update(department.id, { name: name.trim() })
        toast.success('部门已更新')
      } else if (parentId) {
        await departmentApi.create({ name: name.trim(), parentId })
        toast.success('部门已创建')
      }
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title={department ? '编辑部门' : '添加子部门'}
      onClose={onClose}
      contextBar={parentName ? `父部门：${parentName}` : undefined}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={department ? '保存' : '创建'}
          onPrimary={handleSubmit}
          primaryDisabled={submitting}
        />
      }
    >
      <div className="max-w-md space-y-4">
        <div className="space-y-1.5">
          <Label>部门名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              setError('')
              onSetDirty(true)
            }}
            placeholder="输入部门名称"
            autoFocus
          />
          {error && <p className="text-xs text-destructive">{error}</p>}
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}

export function DeptFormWorkflow({ entry, onClose, onSetDirty }: WorkflowComponentProps) {
  const department = entry.payload.department as Department | null | undefined
  const parentId = (entry.payload.parentId as string) ?? department?.parentId ?? ''
  const parentName = (entry.payload.parentName as string) ?? ''
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const formKey = department?.id ?? `new-${parentId}`

  return (
    <DeptFormInner
      key={formKey}
      department={department}
      parentId={parentId}
      parentName={parentName}
      onClose={onClose}
      onSetDirty={onSetDirty}
      onSuccess={onSuccess}
    />
  )
}
