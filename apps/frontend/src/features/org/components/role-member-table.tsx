import { useMemo, useState } from 'react'
import type { Member, Role } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Search, UserPlus, Users } from 'lucide-react'

const DEFAULT_PAGE_SIZE = 10

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

interface RoleMemberTableProps {
  role: Role
  members: Member[]
  onRemoveMember: (member: Member) => void
  onAddMember: () => void
}

export function RoleMemberTable({
  role,
  members,
  onRemoveMember,
  onAddMember,
}: RoleMemberTableProps) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)
  const [pageInputValue, setPageInputValue] = useState('1')

  const filtered = useMemo(
    () => members.filter((m) => m.alias.toLowerCase().includes(search.toLowerCase())),
    [members, search],
  )

  const total = filtered.length
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const safePage = Math.min(page, totalPages)
  const paged = filtered.slice((safePage - 1) * pageSize, safePage * pageSize)

  // Reset to page 1 when search changes
  const handleSearchChange = (value: string) => {
    setSearch(value)
    setPage(1)
    setPageInputValue('1')
  }

  const handlePageChange = (newPage: number) => {
    const clamped = Math.max(1, Math.min(newPage, totalPages))
    setPage(clamped)
    setPageInputValue(String(clamped))
  }

  const handlePageSizeChange = (size: number) => {
    setPageSize(size)
    setPage(1)
    setPageInputValue('1')
  }

  return (
    <div className="flex-1 flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h3 className="text-sm font-semibold text-foreground">{role.name}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {role.type === 'preset' ? '系统预设角色' : '自定义角色'} · {members.length} 名成员
          </p>
        </div>
        <Button size="sm" className="gap-1.5" onClick={onAddMember}>
          <UserPlus className="h-3.5 w-3.5" strokeWidth={1.5} />
          添加成员
        </Button>
      </div>

      {/* Search */}
      <div className="relative mb-4 max-w-xs">
        <Search
          className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground"
          strokeWidth={1.5}
        />
        <Input
          type="text"
          placeholder="搜索成员..."
          value={search}
          onChange={(e) => handleSearchChange(e.target.value)}
          className="pl-8 h-8 text-sm"
        />
      </div>

      {/* Table */}
      <div className="border border-border rounded-lg overflow-hidden flex-1">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wide">
                姓名
              </TableHead>
              <TableHead className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wide">
                角色
              </TableHead>
              <TableHead className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wide text-right">
                操作
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paged.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={3} className="px-4 py-12 text-center">
                  <div className="flex flex-col items-center gap-2">
                    <Users className="h-8 w-8 text-muted-foreground/40" strokeWidth={1.5} />
                    <p className="text-sm text-muted-foreground">暂无成员</p>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              paged.map((member) => (
                <TableRow key={member.id} className="border-border-subtle hover:bg-muted/50">
                  <TableCell className="px-4 py-3 text-sm text-foreground">{member.alias}</TableCell>
                  <TableCell className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {member.roles.map((r) => (
                        <Badge key={r} variant="secondary" className="text-xs font-normal">
                          {r}
                        </Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="px-4 py-3 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs text-destructive hover:text-destructive hover:bg-red-50"
                      onClick={() => onRemoveMember(member)}
                    >
                      移除
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {total > 0 && (
        <div className="flex items-center justify-end gap-4 border-t border-border px-4 py-3 text-sm text-muted-foreground">
          <span>
            共 <span className="tabular-nums font-medium text-foreground">{total}</span> 条
          </span>

          <Pagination className="mx-0 w-auto">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  onClick={() => handlePageChange(safePage - 1)}
                  disabled={safePage <= 1}
                  aria-disabled={safePage <= 1}
                />
              </PaginationItem>

              {generatePageNumbers(safePage, totalPages).map((p, i) => (
                <PaginationItem key={`page-${i}`}>
                  {p === '...' ? (
                    <PaginationEllipsis />
                  ) : (
                    <PaginationLink
                      isActive={p === safePage}
                      onClick={() => handlePageChange(p as number)}
                    >
                      {p}
                    </PaginationLink>
                  )}
                </PaginationItem>
              ))}

              <PaginationItem>
                <PaginationNext
                  onClick={() => handlePageChange(safePage + 1)}
                  disabled={safePage >= totalPages}
                  aria-disabled={safePage >= totalPages}
                />
              </PaginationItem>
            </PaginationContent>
          </Pagination>

          <select
            value={pageSize}
            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
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
                    handlePageChange(clamped)
                  }
                }
              }}
              onBlur={() => {
                const target = parseInt(pageInputValue)
                if (!isNaN(target)) {
                  const clamped = Math.max(1, Math.min(target, totalPages))
                  if (clamped !== safePage) handlePageChange(clamped)
                  setPageInputValue(String(clamped))
                } else {
                  setPageInputValue(String(safePage))
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

// Add member dialog
interface AddMemberDialogProps {
  open: boolean
  roleId: string
  existingMemberIds: string[]
  onAdd: (memberId: string) => void
  onClose: () => void
}

export function AddMemberDialog({
  open,
  existingMemberIds,
  onAdd,
  onClose,
  onSearchMembers,
}: AddMemberDialogProps & {
  onSearchMembers: (keyword: string) => Promise<Member[]>
}) {
  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<Member[]>([])
  const [loading, setLoading] = useState(false)

  const handleClose = () => {
    setKeyword('')
    setResults([])
    onClose()
  }

  const handleSearch = async () => {
    if (!keyword.trim()) return
    setLoading(true)
    try {
      const items = await onSearchMembers(keyword)
      setResults(items.filter((member) => !existingMemberIds.includes(member.id)))
    } catch {
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) handleClose()
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>添加角色成员</DialogTitle>
        </DialogHeader>

        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search
              className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground"
              strokeWidth={1.5}
            />
            <Input
              placeholder="输入姓名搜索..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="pl-8"
            />
          </div>
          <Button onClick={handleSearch} disabled={loading}>
            搜索
          </Button>
        </div>

        <div className="max-h-60 overflow-y-auto border border-border rounded-md">
          {results.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 gap-2">
              <Search className="h-6 w-6 text-muted-foreground/40" strokeWidth={1.5} />
              <p className="text-sm text-muted-foreground">
                {keyword ? '无匹配成员' : '请搜索成员'}
              </p>
            </div>
          ) : (
            <ul className="divide-y divide-border-subtle">
              {results.map((m) => (
                <li
                  key={m.id}
                  className="flex items-center justify-between px-4 py-2.5 hover:bg-muted/50 transition-colors"
                >
                  <div>
                    <span className="text-sm text-foreground">{m.alias}</span>
                    <span className="text-xs text-muted-foreground ml-2">{m.departmentName}</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 text-xs"
                    onClick={() => onAdd(m.id)}
                  >
                    添加
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
