import { useRef } from 'react'
import { Upload } from 'lucide-react'
import type { BatchImportRow } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { WorkflowFormLayout } from '../components/workflow-form-layout'

function parseCsv(text: string): BatchImportRow[] {
  const lines = text.trim().split(/\r?\n/).filter(Boolean)
  if (lines.length < 2) return []
  const rows: BatchImportRow[] = []
  for (let i = 1; i < lines.length; i++) {
    const cols = lines[i].split(',').map((c) => c.trim().replace(/^"|"$/g, ''))
    if (cols.length < 4) continue
    rows.push({
      name: cols[0],
      phone: cols[1],
      email: cols[2],
      departmentName: cols[3],
    })
  }
  return rows
}

export function MemberImportWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps<'member-import'>) {
  const defaultDeptName = (entry.payload.defaultDeptName as string) ?? ''
  const inputRef = useRef<HTMLInputElement>(null)

  const handleFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = () => {
      const text = String(reader.result ?? '')
      const rows = parseCsv(text)
      if (rows.length === 0) return
      onPush('import-preview', {
        rows,
        onSuccess: entry.payload.onSuccess,
      })
    }
    reader.readAsText(file)
  }

  return (
    <WorkflowPanelChrome
      title="批量导入成员"
      onClose={onClose}
      footer={<WorkflowPanelFooter onCancel={onClose} primaryLabel="关闭" onPrimary={onClose} />}
    >
      <WorkflowFormLayout>
        <p className="text-sm text-muted-foreground">
          上传 CSV 文件，列顺序：姓名、手机号、邮箱、部门名称。首行为表头。
        </p>
        {defaultDeptName && (
          <p className="text-sm text-muted-foreground">当前部门：{defaultDeptName}</p>
        )}
        <input
          ref={inputRef}
          type="file"
          accept=".csv,text/csv"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) {
              onSetDirty(true)
              handleFile(file)
            }
            e.target.value = ''
          }}
        />
        <WorkflowInfoBox variant="dropzone" onClick={() => inputRef.current?.click()}>
          <Upload className="h-8 w-8 text-primary" />
          <span className="text-sm font-medium">点击上传 CSV</span>
          <span className="text-xs text-muted-foreground">支持 .csv 格式</span>
        </WorkflowInfoBox>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
