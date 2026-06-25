import { useMemo, useState } from 'react'
import type { Permission } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'

export function PermissionPickerWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'permission-picker'>) {
  const permissions = useMemo(
    () => (entry.payload.permissions as Permission[]) ?? [],
    [entry.payload.permissions],
  )
  const initialSelected = (entry.payload.selected as string[]) ?? []
  const onConfirm = entry.payload.onConfirm as ((perms: string[]) => void) | undefined
  const [selected, setSelected] = useState<string[]>(initialSelected)

  const grouped = useMemo(
    () =>
      permissions.reduce<Record<string, Permission[]>>((acc, p) => {
        if (!acc[p.group]) acc[p.group] = []
        acc[p.group].push(p)
        return acc
      }, {}),
    [permissions],
  )

  const toggle = (permId: string) => {
    setSelected((prev) =>
      prev.includes(permId) ? prev.filter((id) => id !== permId) : [...prev, permId],
    )
    onSetDirty(true)
  }

  const handleConfirm = () => {
    onConfirm?.(selected)
    onPop()
  }

  return (
    <WorkflowPanelChrome
      title="选择权限"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter onCancel={onPop} primaryLabel="确认" onPrimary={handleConfirm} />
      }
    >
      <WorkflowFormLayout variant="full" className="max-h-[60vh] overflow-y-auto">
        {Object.entries(grouped).map(([group, perms], idx) => (
          <div key={group}>
            {idx > 0 && <Separator className="mb-3" />}
            <p className="text-xs font-semibold text-muted-foreground mb-2">{group}</p>
            <div className="space-y-2 pl-1">
              {perms.map((perm) => (
                <div key={perm.id} className="flex items-center gap-2">
                  <Checkbox
                    id={`wf-perm-${perm.id}`}
                    checked={selected.includes(perm.id)}
                    onCheckedChange={() => toggle(perm.id)}
                  />
                  <Label
                    htmlFor={`wf-perm-${perm.id}`}
                    className="text-sm font-normal cursor-pointer"
                  >
                    {perm.name}
                  </Label>
                </div>
              ))}
            </div>
          </div>
        ))}
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
