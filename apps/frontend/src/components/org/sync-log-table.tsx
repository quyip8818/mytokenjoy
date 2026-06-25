import { useEffect, useState } from 'react'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { SyncLog } from '@/api/types'
import { useApis } from '@/api/use-apis'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { PageLoading } from '@/components/ui/page-loading'
import { EmptyState } from '@/components/ui/empty-state'
import { StatusBadge } from '@/components/ui/status-badge'
import { SYNC_RESULT_VARIANTS } from '@/lib/labels'

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
      return (
        <StatusBadge variant={SYNC_RESULT_VARIANTS[value] ?? 'neutral'}>
          {resultLabels[value]}
        </StatusBadge>
      )
    },
  }),
  columnHelper.accessor('detail', { header: '详情' }),
]

export function SyncLogTable() {
  const apis = useApis()
  const [logs, setLogs] = useState<SyncLog[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apis.syncApi.getLogs(1, 10).then((res) => {
      setLogs(res.items)
      setLoading(false)
    })
  }, [apis])

  // eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable function refs
  const table = useReactTable({
    data: logs,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return <PageLoading className="py-8" />
  }

  if (logs.length === 0) {
    return <EmptyState compact title="暂无同步记录" description="执行同步后记录将显示在这里" />
  }

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id} className="hover:bg-transparent">
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
  )
}
