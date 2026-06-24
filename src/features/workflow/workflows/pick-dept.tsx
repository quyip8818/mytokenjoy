import { useEffect, useState } from 'react'
import { departmentApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { flattenDepartments } from '@/lib/org'

export function PickDeptWorkflow({ entry, onPop, onClose, onSetDirty }: WorkflowComponentProps) {
  const selectedId = (entry.payload.selectedId as string) ?? ''
  const onConfirm = entry.payload.onConfirm as ((deptId: string) => void) | undefined
  const [departments, setDepartments] = useState<{ id: string; name: string; level: number }[]>([])
  const [search, setSearch] = useState('')
  const [picked, setPicked] = useState(selectedId)

  useEffect(() => {
    departmentApi.getTree().then((tree) => setDepartments(flattenDepartments(tree)))
  }, [])

  const filtered = departments.filter((d) => d.name.toLowerCase().includes(search.toLowerCase()))

  const handleConfirm = () => {
    if (!picked) return
    onConfirm?.(picked)
    onPop()
  }

  return (
    <WorkflowPanelChrome
      title="选择部门"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel="确认"
          onPrimary={handleConfirm}
          primaryDisabled={!picked}
        />
      }
    >
      <div className="space-y-4">
        <Input
          placeholder="搜索部门..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="space-y-1 max-h-[60vh] overflow-y-auto">
          {filtered.map((d) => (
            <button
              key={d.id}
              type="button"
              onClick={() => {
                setPicked(d.id)
                onSetDirty(true)
              }}
              className={cn(
                'w-full text-left rounded-lg border border-border/50 px-4 py-3 text-sm hover:bg-indigo-50/30',
                picked === d.id && 'border-indigo-200 bg-indigo-50/40',
              )}
            >
              {'　'.repeat(d.level)}
              {d.name}
            </button>
          ))}
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
