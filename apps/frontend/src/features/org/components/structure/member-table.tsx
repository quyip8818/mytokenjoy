import { useMemo, useState } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type RowSelectionState,
} from '@tanstack/react-table'
import type { Member } from '@/api/types'
import { orgApi } from '@/api/org'
import { useSession } from '@/features/session'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

interface MemberTableProps {
  data: Member[]
  total: number
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  onEdit: (member: Member) => void
  onStatusChange: (ids: string[], status: 'active' | 'disabled') => void
  onDelete: (ids: string[]) => void
  rowSelection: RowSelectionState
  onRowSelectionChange: (selection: RowSelectionState) => void
}

const statusConfig: Record<string, { label: string; className: string }> = {
  active: { label: '已激活', className: 'bg-emerald-50 text-emerald-700' },
  disabled: { label: '已停用', className: 'bg-slate-100 text-slate-600' },
  pending: { label: '待激活', className: 'bg-amber-50 text-amber-700' },
}

const sourceLabels: Record<string, string> = {
  manual: '手动',
  csv: 'CSV',
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企微',
  invited: '邀请',
  imported: '导入',
}

const registrationChannelLabels: Record<string, string> = {
  sms: '短信',
  email: '邮件',
  admin_link: '管理员链接',
}

const columnHelper = createColumnHelper<Member>()

function generatePageNumbers(current: number, total: number): (number | '...')[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | '...')[] = [1]
  let left = Math.max(2, current - 2)
  let right = Math.min(total - 1, current + 2)
  if (current <= 4) {
    left = 2
    right = 5
  } else if (current >= total - 3) {
    left = total - 4
    right = total - 1
  }
  if (left > 2) pages.push('...')
  for (let i = left; i <= right; i++) pages.push(i)
  if (right < total - 1) pages.push('...')
  pages.push(total)
  return pages
}

export function MemberTable({
  data,
  total,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  onEdit,
  onStatusChange,
  onDelete,
  rowSelection,
  onRowSelectionChange,
}: MemberTableProps) {
  const { memberId: currentMemberId, member: currentMember } = useSession()
  const columns = useMemo(() => {
    const currentMemberRoles = currentMember?.roles ?? []
    return [
      columnHelper.display({
        id: 'select',
        header: ({ table }) => (
          <Checkbox
            checked={table.getIsAllPageRowsSelected()}
            onCheckedChange={(c) => table.toggleAllPageRowsSelected(!!c)}
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            disabled={!row.getCanSelect()}
            onCheckedChange={(c) => row.toggleSelected(!!c)}
          />
        ),
        size: 40,
      }),
      columnHelper.accessor('name', {
        header: '姓名',
        cell: (info) => (
          <span className="font-medium text-foreground">{info.getValue() || '—'}</span>
        ),
      }),
      columnHelper.accessor('alias', {
        header: '昵称',
        cell: (info) => <span className="text-muted-foreground">{info.getValue() || '—'}</span>,
      }),
      columnHelper.accessor('departmentName', { header: '部门' }),
      columnHelper.accessor('jobTitle', {
        header: '职位',
        cell: (info) => <span className="text-muted-foreground">{info.getValue() || '—'}</span>,
      }),
      columnHelper.accessor('status', {
        header: '状态',
        cell: (info) => {
          const s = statusConfig[info.getValue()] ?? statusConfig.active
          return (
            <span
              className={cn(
                'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                s.className,
              )}
            >
              {s.label}
            </span>
          )
        },
      }),
      columnHelper.accessor('source', {
        header: '来源',
        cell: (info) => (
          <span className="text-muted-foreground text-xs">
            {sourceLabels[info.getValue()] ?? info.getValue()}
          </span>
        ),
      }),
      columnHelper.accessor('registrationChannel', {
        header: '注册方式',
        cell: (info) => {
          const val = info.getValue()
          if (!val) return <span className="text-muted-foreground text-xs">未注册</span>
          return (
            <span className="text-muted-foreground text-xs">
              {registrationChannelLabels[val] ?? val}
            </span>
          )
        },
      }),
      columnHelper.display({
        id: 'actions',
        header: '操作',
        cell: ({ row }) => {
          const member = row.original
          const canInvite =
            currentMemberRoles.includes('超级管理员') || currentMemberRoles.includes('组织管理员')
          const handleGetInviteLink = async () => {
            try {
              const { inviteLink } = await orgApi.members.getInviteLink(member.id)
              await navigator.clipboard.writeText(inviteLink)
              toast.success('邀请链接已复制到剪贴板')
            } catch {
              toast.error('获取邀请链接失败')
            }
          }
          const handleResendInvite = async () => {
            try {
              await orgApi.members.batchInvite([member.id])
              toast.success('邀请已重新发送')
            } catch {
              toast.error('重新发送邀请失败')
            }
          }
          return (
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="xs" onClick={() => onEdit(member)}>
                编辑
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon-xs" aria-label="更多操作">
                    <MoreHorizontal className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {canInvite && member.status === 'pending' && (
                    <DropdownMenuItem onClick={handleGetInviteLink}>获取邀请链接</DropdownMenuItem>
                  )}
                  {member.status === 'pending' && (
                    <DropdownMenuItem onClick={handleResendInvite}>重新发送邀请</DropdownMenuItem>
                  )}
                  {member.status === 'active' && (
                    <DropdownMenuItem onClick={() => onStatusChange([member.id], 'disabled')}>
                      停用
                    </DropdownMenuItem>
                  )}
                  {member.status === 'disabled' && (
                    <DropdownMenuItem onClick={() => onStatusChange([member.id], 'active')}>
                      启用
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuItem
                    className="text-destructive"
                    onClick={() => onDelete([member.id])}
                  >
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          )
        },
      }),
    ]
  }, [onEdit, onStatusChange, onDelete, currentMember])

  const table = useReactTable({
    data,
    columns,
    state: { rowSelection },
    enableRowSelection: (row) => row.original.id !== currentMemberId,
    onRowSelectionChange: (updater) => {
      onRowSelectionChange(typeof updater === 'function' ? updater(rowSelection) : updater)
    },
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.id,
    manualPagination: true,
    rowCount: total,
  })

  const totalPages = Math.ceil(total / pageSize)
  const [pageInputValue, setPageInputValue] = useState(String(page))

  // Sync input when page changes externally
  if (
    pageInputValue !== String(page) &&
    document.activeElement?.getAttribute('aria-label') !== '跳转页码'
  ) {
    setPageInputValue(String(page))
  }

  return (
    <div className="flex flex-1 flex-col overflow-hidden rounded-md border border-border">
      <div className="flex-1 overflow-auto">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((hg) => (
              <TableRow key={hg.id} className="bg-muted hover:bg-muted">
                {hg.headers.map((h) => (
                  <TableHead key={h.id}>
                    {h.isPlaceholder ? null : flexRender(h.column.columnDef.header, h.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                data-state={row.getIsSelected() && 'selected'}
                className="even:bg-muted/40 hover:bg-muted/50 transition-colors duration-100"
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
            {data.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="py-12 text-center text-sm text-muted-foreground"
                >
                  暂无成员数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {total > 0 && (
        <div className="flex items-center justify-end gap-4 border-t border-border px-4 py-3 text-sm text-muted-foreground">
          <span>
            共 <span className="tabular-nums font-medium text-foreground">{total}</span> 条
          </span>

          <Pagination className="mx-0 w-auto">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  onClick={() => onPageChange(page - 1)}
                  disabled={page <= 1}
                  aria-disabled={page <= 1}
                />
              </PaginationItem>

              {generatePageNumbers(page, totalPages).map((p, i) => (
                <PaginationItem key={`page-${i}`}>
                  {p === '...' ? (
                    <PaginationEllipsis />
                  ) : (
                    <PaginationLink isActive={p === page} onClick={() => onPageChange(p as number)}>
                      {p}
                    </PaginationLink>
                  )}
                </PaginationItem>
              ))}

              <PaginationItem>
                <PaginationNext
                  onClick={() => onPageChange(page + 1)}
                  disabled={page >= totalPages}
                  aria-disabled={page >= totalPages}
                />
              </PaginationItem>
            </PaginationContent>
          </Pagination>

          <select
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
            className="h-8 rounded-md border border-input bg-background px-2 text-sm outline-none focus:ring-1 focus:ring-ring"
          >
            <option value={10}>10 条/页</option>
            <option value={20}>20 条/页</option>
            <option value={50}>50 条/页</option>
            <option value={100}>100 条/页</option>
          </select>

          <div className="flex items-center gap-1">
            <span>跳至</span>
            <input
              type="number"
              min={1}
              max={totalPages}
              value={pageInputValue}
              onChange={(e) => setPageInputValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  const target = parseInt(pageInputValue)
                  if (!isNaN(target)) {
                    const clamped = Math.max(1, Math.min(target, totalPages))
                    onPageChange(clamped)
                    setPageInputValue(String(clamped))
                  }
                }
              }}
              onBlur={() => {
                const target = parseInt(pageInputValue)
                if (!isNaN(target)) {
                  const clamped = Math.max(1, Math.min(target, totalPages))
                  if (clamped !== page) onPageChange(clamped)
                  setPageInputValue(String(clamped))
                } else {
                  setPageInputValue(String(page))
                }
              }}
              className="h-8 w-12 rounded-md border border-input bg-background px-1 text-center text-sm tabular-nums outline-none focus:ring-1 focus:ring-ring [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
              aria-label="跳转页码"
            />
            <span>页</span>
          </div>
        </div>
      )}
    </div>
  )
}
