import { useEffect, useState } from 'react'
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table'
import type { SyncLog } from '@/api/types'
import { syncApi } from '@/api/org'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

const columnHelper = createColumnHelper<SyncLog>()

const typeLabels: Record<SyncLog['type'], string> = {
  scheduled: '定时',
  manual: '手动',
}

const resultLabels: Record<SyncLog['result'], string> = {
  success: '成功',
  partial_failure: '部分失败',
  failure: '失败',
}

const resultBadge: Record<SyncLog['result'], { variant: 'default' | 'secondary' | 'destructive'; className?: string }> = {
  success: { variant: 'default', className: 'bg-green-100 text-green-700 hover:bg-green-100' },
  partial_failure: { variant: 'secondary' },
  failure: { variant: 'destructive' },
}

const columns = [
  columnHelper.accessor('time', { header: '时间' }),
  columnHelper.accessor('type', {
    header: '类型',
    cell: (info) => typeLabels[info.getValue()],
  }),
  columnHelper.accessor('result', {
    header: '结果',
    cell: (info) => {
      const value = info.getValue()
      const { variant, className } = resultBadge[value]
      return <Badge variant={variant} className={className}>{resultLabels[value]}</Badge>
    },
  }),
  columnHelper.accessor('detail', { header: '详情' }),
]

export function SyncLogTable() {
  const [logs, setLogs] = useState<SyncLog[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    syncApi.getLogs(1, 10).then((res) => {
      setLogs(res.items)
      setLoading(false)
    })
  }, [])

  const table = useReactTable({
    data: logs,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return <p className="text-sm text-muted-foreground">加载中...</p>
  }

  if (logs.length === 0) {
    return <p className="text-sm text-muted-foreground">暂无同步记录</p>
  }

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id}>
                  {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
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
  )
}
