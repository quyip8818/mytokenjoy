import { useState } from 'react'
import type { BudgetNode, ProjectView, Role } from '@/api/types'
import type { AlertRuleView } from '@/features/budget'
import { FormDialog } from '@/components/ui/form-dialog'
import { useAlertRuleForm } from '../hooks/use-alert-rule-form'
import { AlertRuleFormFields } from './alert-rule-form-fields'

interface AlertRuleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  rule: AlertRuleView | null
  tree: BudgetNode[]
  projects: ProjectView[]
  roles: Role[]
  onSave: (view: AlertRuleView, existingId?: string) => Promise<void>
}

export function AlertRuleDialog({
  open,
  onOpenChange,
  rule,
  tree,
  projects,
  roles,
  onSave,
}: AlertRuleDialogProps) {
  if (!open) return null

  return (
    <AlertRuleDialogInner
      open={open}
      onOpenChange={onOpenChange}
      rule={rule}
      tree={tree}
      projects={projects}
      roles={roles}
      onSave={onSave}
    />
  )
}

function AlertRuleDialogInner({
  open,
  onOpenChange,
  rule,
  tree,
  projects,
  roles,
  onSave,
}: AlertRuleDialogProps) {
  const form = useAlertRuleForm(rule)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSave() {
    const validationError = form.validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setError(null)
    setSaving(true)
    try {
      await onSave(form.buildPayload(), rule?.id)
      onOpenChange(false)
    } catch {
      setError('保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={rule ? '编辑预警规则' : '创建预警规则'}
      error={error}
      busy={saving}
      submitLabel={rule ? '保存' : '创建'}
      onSubmit={handleSave}
      className="sm:max-w-lg"
    >
      <AlertRuleFormFields tree={tree} projects={projects} roles={roles} form={form} />
    </FormDialog>
  )
}
