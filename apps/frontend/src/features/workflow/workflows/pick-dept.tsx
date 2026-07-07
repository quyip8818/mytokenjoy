import { useEffect, useState } from 'react'
import { useApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { WorkflowListItem, WorkflowScrollList } from '../components/workflow-list-item'
import { WorkflowPickerShell } from '../components/workflow-picker-shell'
import { Input } from '@/components/ui/input'
import { flattenDepartments } from '@/features/org/lib/departments'

export function PickDeptWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'pick-dept'>) {
  const apis = useApis()
  const selectedId = (entry.payload.selectedId as string) ?? ''
  const onConfirm = entry.payload.onConfirm as ((deptId: string) => void) | undefined
  const [departments, setDepartments] = useState<{ id: string; name: string; level: number }[]>([])
  const [search, setSearch] = useState('')
  const [picked, setPicked] = useState(selectedId)

  useEffect(() => {
    apis.departmentApi.getTree().then((tree) => setDepartments(flattenDepartments(tree)))
  }, [apis])

  const filtered = departments.filter((d) => d.name.toLowerCase().includes(search.toLowerCase()))

  const handleConfirm = () => {
    if (!picked) return
    onConfirm?.(picked)
    onPop()
  }

  return (
    <WorkflowPickerShell
      title="选择部门"
      onPop={onPop}
      onClose={onClose}
      onConfirm={handleConfirm}
      primaryDisabled={!picked}
    >
      <WorkflowFormLayout variant="full">
        <Input
          placeholder="搜索部门..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <WorkflowScrollList>
          {filtered.map((d) => (
            <WorkflowListItem
              key={d.id}
              selected={picked === d.id}
              onClick={() => {
                setPicked(d.id)
                onSetDirty(true)
              }}
            >
              {'　'.repeat(d.level)}
              {d.name}
            </WorkflowListItem>
          ))}
        </WorkflowScrollList>
      </WorkflowFormLayout>
    </WorkflowPickerShell>
  )
}
