import { useMemo } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type RowSelectionState,
} from '@tanstack/react-table'
import type { Member } from '@/api/types'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface MemberTableProps {
  data: Member[]
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  onEdit: (member: Member) => void
  onStatusChange: (ids: string[], status: 'active' | 'inactive') => void
  onDelete: (ids: string[]) => void
  rowSelection: RowSelectionState
  onRowSelectionChange: (selection: RowSelectionState) => void
}

function maskPhone(phone: string): string {
  if (phone.length >= 7) {
    return phone.slice(0, 3) + '****' + phone.slice(-4)
  }
  return phone
}

const statusMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  active: { label: '已激活', variant: 'default' },
  inactive: { label: '已停用', variant: 'secondary' },
  pending: { label: '待激活', variant: 'outline' },
}

const columnHelper = createColumnHelper<Member>()

export function MemberTable({
  data,
  total,
  page,
  pageSize,
  onPageChange,
  onEdit,
  onStatusChange,
  onDelete,
  rowSelection,
  onRowSelectionChange,
}: MemberTableProps) {
  const columns = useMemo(() => [
    columnHelper.display({
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(checked) => table.toggleAllPageRowsSelected(!!checked)}
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(checked) => row.toggleSelected(!!checked)}
        />
      ),
      size: 40,
    }),
    columnHelper.accessor('name', {
      header: '姓名',
      cell: info => <span className="font-medium text-foreground">{info.getValue()}</span>,
    }),
    columnHelper.accessor('departmentName', {
      header: '部门',
    }),
    columnHelper.accessor('phone', {
      header: '手机号',
      cell: info => maskPhone(info.getValue()),
    }),
    columnHelper.accessor('status', {
      header: '状态',
      cell: info => {
        const s = statusMap[info.getValue()] ?? statusMap.active
        return <Badge variant={s.variant}>{s.label}</Badge>
      },
    }),
    columnHelper.display({
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const member = row.original
        return (
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="xs" onClick={() => onEdit(member)}>
              编辑
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="xs">
                更多
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {member.status === 'active' ? (
                  <DropdownMenuItem onClick={() => onStatusChange([member.id], 'inactive')}>
                    停用
                  </DropdownMenuItem>
                ) : (
                  <DropdownMenuItem onClick={() => onStatusChange([member.id], 'active')}>
                    启用
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem variant="destructive" onClick={() => onDelete([member.id])}>
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        )
      },
    }),
  ], [onEdit, onStatusChange, onDelete])

  const table = useReactTable({
    data,
    columns,
    state: { rowSelection },
    onRowSelectionChange: updater => {
      const next = typeof updater === 'function' ? updater(rowSelection) : updater
      onRowSelectionChange(next)
    },
    getCoreRowModel: getCoreRowModel(),
    getRowId: row => row.id,
    manualPagination: true,
    rowCount: total,
  })

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div>
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map(headerGroup => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map(header => (
                <TableHead key={header.id}>
                  {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.map(row => (
            <TableRow key={row.id}>
              {row.getVisibleCells().map(cell => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
          {data.length === 0 && (
            <TableRow>
              <TableCell colSpan={columns.length} className="text-center text-muted-foreground py-8">
                暂无数据
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>

      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4 text-sm text-muted-foreground">
          <span>共 {total} 条</span>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
            >上一页</Button>
            <span className="px-3 py-1">{page} / {totalPages}</span>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
            >下一页</Button>
          </div>
        </div>
      )}
    </div>
  )
}
