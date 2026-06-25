import { useState } from 'react'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { ImportFailure, ImportResult } from '@/api/types'
import { dataSourceApi } from '@/api/org'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface ImportResultProps {
  result: ImportResult
  onNavigateOrg?: () => void
  onUpdate: (result: ImportResult) => void
}

const columnHelper = createColumnHelper<ImportFailure>()

const columns = [
  columnHelper.accessor('name', { header: '姓名' }),
  columnHelper.accessor('employeeId', { header: '工号' }),
  columnHelper.accessor('reason', { header: '失败原因' }),
]

export function ImportResultView({ result, onNavigateOrg, onUpdate }: ImportResultProps) {
  const [retrying, setRetrying] = useState<Set<string>>(new Set())
  const [retryingAll, setRetryingAll] = useState(false)

  // eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable function refs
  const table = useReactTable({
    data: result.failures,
    columns: [
      ...columns,
      columnHelper.display({
        id: 'actions',
        header: '操作',
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="sm"
            disabled={retrying.has(row.original.id)}
            onClick={() => handleRetry(row.original.id)}
          >
            {retrying.has(row.original.id) ? '重试中...' : '重试'}
          </Button>
        ),
      }),
    ],
    getCoreRowModel: getCoreRowModel(),
  })

  const handleRetry = async (id: string) => {
    setRetrying((prev) => new Set(prev).add(id))
    try {
      const res = await dataSourceApi.retryImport([id])
      const updatedFailures = result.failures.filter((f) => f.id !== id)
      onUpdate({
        successMembers: result.successMembers + res.successMembers,
        successDepartments: result.successDepartments + res.successDepartments,
        failures: [...updatedFailures, ...res.failures],
      })
    } finally {
      setRetrying((prev) => {
        const next = new Set(prev)
        next.delete(id)
        return next
      })
    }
  }

  const handleRetryAll = async () => {
    const ids = result.failures.map((f) => f.id)
    setRetryingAll(true)
    try {
      const res = await dataSourceApi.retryImport(ids)
      onUpdate({
        successMembers: result.successMembers + res.successMembers,
        successDepartments: result.successDepartments + res.successDepartments,
        failures: res.failures,
      })
    } finally {
      setRetryingAll(false)
    }
  }

  const allSuccess = result.failures.length === 0

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4 text-sm">
        <span className="text-green-700">
          成功导入：{result.successMembers} 人，{result.successDepartments} 个部门
        </span>
        {result.failures.length > 0 && (
          <span className="text-destructive">失败：{result.failures.length} 条</span>
        )}
      </div>

      {allSuccess && onNavigateOrg && (
        <div className="flex items-center gap-3 p-3 bg-green-50 border border-green-200 rounded-md">
          <span className="text-sm text-green-800">全部导入成功！</span>
          <Button variant="default" size="sm" onClick={onNavigateOrg}>
            跳转到组织架构页面
          </Button>
        </div>
      )}

      {result.failures.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-sm font-medium text-muted-foreground">失败详情</h4>
            <Button variant="destructive" size="sm" onClick={handleRetryAll} disabled={retryingAll}>
              {retryingAll ? '重试中...' : '全部重试失败'}
            </Button>
          </div>
          <div className="overflow-x-auto rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows.map((row) => (
                  <TableRow key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}
    </div>
  )
}
