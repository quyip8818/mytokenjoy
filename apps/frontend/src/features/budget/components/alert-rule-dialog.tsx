import { useState } from 'react'
import type { BudgetNode, BudgetProjectView, Role } from '@/api/types'
import type { AlertRuleView } from '@/features/budget'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useAlertRuleForm } from '../hooks/use-alert-rule-form'
import { AlertRuleFormFields } from './alert-rule-form-fields'

interface AlertRuleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  rule: AlertRuleView | null
  tree: BudgetNode[]
  projects: BudgetProjectView[]
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
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open && (
        <AlertRuleDialogContent
          rule={rule}
          tree={tree}
          projects={projects}
          roles={roles}
          onOpenChange={onOpenChange}
          onSave={onSave}
        />
      )}
    </Dialog>
  )
}

function AlertRuleDialogContent({
  rule,
  tree,
  projects,
  roles,
  onOpenChange,
  onSave,
}: {
  rule: AlertRuleView | null
  tree: BudgetNode[]
  projects: BudgetProjectView[]
  roles: Role[]
  onOpenChange: (open: boolean) => void
  onSave: (view: AlertRuleView, existingId?: string) => Promise<void>
}) {
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
    <DialogContent className="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{rule ? '编辑预警规则' : '创建预警规则'}</DialogTitle>
      </DialogHeader>

      <AlertRuleFormFields tree={tree} projects={projects} roles={roles} form={form} />

      {error && <p className="text-xs text-red-600">{error}</p>}

      <DialogFooter>
        <Button variant="outline" size="sm" onClick={() => onOpenChange(false)} disabled={saving}>
          取消
        </Button>
        <Button size="sm" onClick={handleSave} disabled={saving}>
          {saving ? '保存中…' : rule ? '保存' : '创建'}
        </Button>
      </DialogFooter>
    </DialogContent>
  )
}
