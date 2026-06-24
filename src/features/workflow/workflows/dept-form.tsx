import { useState } from 'react'
import type { Department } from '@/api/types'
import { departmentApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { WorkflowFormField } from '../components/workflow-form-field'
import { useWorkflowSubmit } from '../use-workflow-submit'
import { Input } from '@/components/ui/input'

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
  const [name, setName] = useState(department?.name ?? '')
  const [error, setError] = useState('')
  const { submit, submitting } = useWorkflowSubmit({
    validate: () => (!name.trim() ? '请输入部门名称' : null),
    onSubmit: async () => {
      if (department) {
        await departmentApi.update(department.id, { name: name.trim() })
      } else if (parentId) {
        await departmentApi.create({ name: name.trim(), parentId })
      }
    },
    successMessage: department ? '部门已更新' : '部门已创建',
    onSuccess,
  })

  const handleSubmit = async () => {
    const result = await submit()
    if (!result.ok && result.error) setError(result.error)
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
      <WorkflowFormLayout>
        <WorkflowFormField label="部门名称" error={error}>
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
        </WorkflowFormField>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

export function DeptFormWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'dept-form'>) {
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
