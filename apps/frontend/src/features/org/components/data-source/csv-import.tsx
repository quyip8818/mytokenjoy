import { useCallback, useRef, useState } from 'react'
import { toast } from '@/lib/toast'
import { ArrowLeft, Download, Upload, FileSpreadsheet, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { BatchImportRow, MemberBatchImportResult } from '@/api/types'

const CSV_COLUMNS = [
  { key: 'name', label: '姓名', required: true },
  { key: 'employeeId', label: '工号', required: true },
  { key: 'departmentName', label: '部门', required: true },
  { key: 'phone', label: '手机号', required: false },
  { key: 'email', label: '邮箱', required: false },
  { key: 'alias', label: '昵称', required: false },
  { key: 'jobTitle', label: '职位', required: false },
  { key: 'hireDate', label: '入职时间', required: false },
] as const

type CsvColumnKey = (typeof CSV_COLUMNS)[number]['key']

const EXAMPLE_ROW: Record<CsvColumnKey, string> = {
  name: '张三',
  employeeId: 'EMP001',
  departmentName: '技术部',
  phone: '13800138000',
  email: 'zhangsan@example.com',
  alias: '三哥',
  jobTitle: '高级工程师',
  hireDate: '2024-03-15',
}

interface CsvImportProps {
  onImport: (rows: BatchImportRow[]) => Promise<MemberBatchImportResult>
  onBack: () => void
}

interface ValidationError {
  row: number
  errors: string[]
}

export function CsvImport({ onImport, onBack }: CsvImportProps) {
  const [rows, setRows] = useState<BatchImportRow[]>([])
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [importing, setImporting] = useState(false)
  const [fileName, setFileName] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const downloadExample = useCallback(() => {
    const header = CSV_COLUMNS.map((c) => c.label).join(',')
    const row = CSV_COLUMNS.map((c) => EXAMPLE_ROW[c.key]).join(',')
    const bom = '﻿'
    const blob = new Blob([bom + header + '\n' + row + '\n'], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'example.csv'
    a.click()
    URL.revokeObjectURL(url)
  }, [])

  const parseCsv = useCallback((text: string): BatchImportRow[] => {
    const lines = text.split(/\r?\n/).filter((l) => l.trim())
    if (lines.length < 2) return []
    const headers = lines[0].split(',').map((h) => h.trim())
    const colIndex: Partial<Record<CsvColumnKey, number>> = {}
    for (const col of CSV_COLUMNS) {
      const idx = headers.indexOf(col.label)
      if (idx !== -1) colIndex[col.key] = idx
    }
    return lines.slice(1).map((line) => {
      const cells = line.split(',').map((c) => c.trim())
      const row: BatchImportRow = {
        name: '',
        phone: '',
        email: '',
        departmentName: '',
        employeeId: '',
        alias: '',
        jobTitle: '',
        hireDate: '',
      }
      for (const col of CSV_COLUMNS) {
        const idx = colIndex[col.key]
        if (idx !== undefined)
          (row as unknown as Record<string, string>)[col.key] = cells[idx] ?? ''
      }
      return row
    })
  }, [])

  const validate = useCallback((parsed: BatchImportRow[]): ValidationError[] => {
    const errors: ValidationError[] = []
    for (let i = 0; i < parsed.length; i++) {
      const row = parsed[i]
      const rowErrors: string[] = []
      if (!row.name.trim()) rowErrors.push('姓名为必填')
      if (!row.employeeId.trim()) rowErrors.push('工号为必填')
      if (!row.departmentName.trim()) rowErrors.push('部门为必填')
      if (!row.phone.trim() && !row.email.trim()) rowErrors.push('手机号或邮箱至少填写一项')
      if (row.hireDate.trim() && !/^\d{4}[-/]\d{1,2}[-/]\d{1,2}$/.test(row.hireDate.trim())) {
        rowErrors.push('入职时间格式应为 YYYY-MM-DD 或 YYYY/M/D')
      }
      if (rowErrors.length > 0) errors.push({ row: i + 1, errors: rowErrors })
    }
    return errors
  }, [])

  const handleFile = useCallback(
    (file: File) => {
      setFileName(file.name)
      const reader = new FileReader()
      reader.onload = (e) => {
        const text = e.target?.result as string
        const parsed = parseCsv(text)
        if (parsed.length === 0) {
          toast.error('CSV 文件为空或格式不正确')
          return
        }
        const errs = validate(parsed)
        setRows(parsed)
        setValidationErrors(errs)
      }
      reader.readAsText(file, 'utf-8')
    },
    [parseCsv, validate],
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      const file = e.dataTransfer.files[0]
      if (file && file.name.endsWith('.csv')) handleFile(file)
      else toast.error('请上传 .csv 文件')
    },
    [handleFile],
  )

  const handleImport = async () => {
    if (validationErrors.length > 0) {
      toast.error('请先修正校验错误后再导入')
      return
    }
    setImporting(true)
    try {
      const result = await onImport(rows)
      if (result.failures.length === 0) {
        toast.success(`成功导入 ${result.imported} 名成员`)
      } else {
        toast.success(`导入 ${result.imported} 名成员`)
        toast.warning(
          `${result.failures.length} 行导入失败：${result.failures
            .map((f) => (f.row > 0 ? `第${f.row}行: ${f.reason}` : f.reason))
            .slice(0, 3)
            .join('；')}`,
          { duration: 8000 },
        )
      }
      onBack()
    } catch {
      toast.error('导入失败，请重试')
    } finally {
      setImporting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" onClick={onBack}>
          <ArrowLeft className="size-3.5" />
          返回
        </Button>
        <h2 className="text-base font-semibold text-foreground">CSV 批量导入成员</h2>
      </div>

      <div className="rounded-lg border border-dashed border-border bg-muted/30 p-6">
        <div className="flex flex-col items-center gap-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <FileSpreadsheet className="size-4" />
            <span>请先下载模板，填写成员信息后上传</span>
          </div>
          <Button variant="outline" onClick={downloadExample}>
            <Download className="size-3.5" />
            下载 CSV 模板
          </Button>
        </div>
      </div>

      <div
        className={cn(
          'flex flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 transition-colors',
          'cursor-pointer hover:border-primary/40 hover:bg-muted/20',
          fileName ? 'border-primary/60 bg-primary/5' : 'border-border',
        )}
        onDragOver={(e) => e.preventDefault()}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        role="button"
        tabIndex={0}
        aria-label="上传 CSV 文件"
      >
        <Upload className="size-6 text-muted-foreground" />
        {fileName ? (
          <p className="text-sm font-medium text-foreground">{fileName}</p>
        ) : (
          <p className="text-sm text-muted-foreground">拖拽 CSV 文件到此处，或点击选择文件</p>
        )}
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) handleFile(file)
          }}
        />
      </div>

      {rows.length > 0 && (
        <>
          {validationErrors.length > 0 && (
            <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4">
              <div className="flex items-center gap-2 text-sm font-medium text-destructive">
                <AlertCircle className="size-4" />
                校验错误（{validationErrors.length} 行）
              </div>
              <ul className="mt-2 space-y-1 text-xs text-destructive">
                {validationErrors.slice(0, 10).map((ve) => (
                  <li key={ve.row}>
                    第 {ve.row} 行：{ve.errors.join('，')}
                  </li>
                ))}
                {validationErrors.length > 10 && (
                  <li>...还有 {validationErrors.length - 10} 行错误</li>
                )}
              </ul>
            </div>
          )}

          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">
              预览（共 {rows.length} 行{validationErrors.length === 0 ? '，校验通过' : ''}）
            </p>
            <div className="overflow-x-auto rounded-lg border border-border">
              <table className="w-full text-xs">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="whitespace-nowrap px-3 py-2 text-left font-medium text-muted-foreground">
                      #
                    </th>
                    {CSV_COLUMNS.map((c) => (
                      <th
                        key={c.key}
                        className="whitespace-nowrap px-3 py-2 text-left font-medium text-muted-foreground"
                      >
                        {c.label}
                        {c.required && <span className="text-destructive">*</span>}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {rows.slice(0, 5).map((row, i) => (
                    <tr key={i} className="border-t border-border">
                      <td className="px-3 py-1.5 text-muted-foreground">{i + 1}</td>
                      {CSV_COLUMNS.map((c) => (
                        <td key={c.key} className="px-3 py-1.5">
                          {(row as unknown as Record<string, string>)[c.key] || '-'}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
              {rows.length > 5 && (
                <p className="border-t border-border px-3 py-2 text-xs text-muted-foreground">
                  ...还有 {rows.length - 5} 行
                </p>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <Button
              variant="outline"

              onClick={() => {
                setRows([])
                setFileName(null)
                setValidationErrors([])
              }}
            >
              重新选择
            </Button>
            <Button disabled={validationErrors.length > 0 || importing} onClick={handleImport}>
              {importing ? '导入中...' : `确认导入 ${rows.length} 名成员`}
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
