import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { SyncLog } from '@/api/types'
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
import { SYNC_RESULT_LABELS, SYNC_RESULT_VARIANTS, SYNC_TYPE_LABELS } from '@/lib/labels'

const columnHelper = createColumnHelper<SyncLog>()

const columns = [
  columnHelper.accessor('time', { header: '时间' }),
  columnHelper.accessor('type', {
    header: '类型',
    cell: (info) => SYNC_TYPE_LABELS[info.getValue()],
  }),
  columnHelper.accessor('result', {
    header: '结果',
    cell: (info) => {
      const value = info.getValue()
      return (
        <StatusBadge variant={SYNC_RESULT_VARIANTS[value] ?? 'neutral'}>
          {SYNC_RESULT_LABELS[value]}
        </StatusBadge>
      )
    },
  }),
  columnHelper.accessor('detail', { header: '详情' }),
]

interface SyncLogTableProps {
  logs: SyncLog[]
  loading: boolean
}

export function SyncLogTable({ logs, loading }: SyncLogTableProps) {
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
