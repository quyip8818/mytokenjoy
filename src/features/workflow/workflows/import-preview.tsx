import { useState } from 'react'
import { toast } from 'sonner'
import type { BatchImportRow } from '@/api/types'
import { memberApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useWorkflow } from '../use-workflow'

export function ImportPreviewWorkflow({ entry, onPop, onClose }: WorkflowComponentProps) {
  const { closeAll } = useWorkflow()
  const rows = (entry.payload.rows as BatchImportRow[]) ?? []
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [submitting, setSubmitting] = useState(false)

  const handleConfirm = async () => {
    setSubmitting(true)
    try {
      const result = await memberApi.batchImport(rows)
      toast.success(`成功导入 ${result.imported} 名成员`)
      if (result.failures.length > 0) {
        toast.warning(`${result.failures.length} 行导入失败`)
      }
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('导入失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="导入预览"
      showBack
      onBack={onPop}
      onClose={onClose}
      contextBar={`共 ${rows.length} 条记录`}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel="确认导入"
          onPrimary={handleConfirm}
          primaryDisabled={rows.length === 0 || submitting}
        />
      }
    >
      <div className="max-h-[55vh] overflow-auto rounded-lg border border-border/60">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>姓名</TableHead>
              <TableHead>手机号</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>部门</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, i) => (
              <TableRow key={i}>
                <TableCell>{row.name}</TableCell>
                <TableCell>{row.phone}</TableCell>
                <TableCell>{row.email}</TableCell>
                <TableCell>{row.departmentName}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </WorkflowPanelChrome>
  )
}
