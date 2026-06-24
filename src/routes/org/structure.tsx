import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router'
import { toast } from 'sonner'
import type { Department, Member, Paginated } from '@/api/types'
import { departmentApi, memberApi } from '@/api/org'
import { approvalApi } from '@/api/keys'
import { DepartmentTree } from '@/components/org/department-tree'
import { MemberTable } from '@/components/org/member-table'
import { EmptyState } from '@/components/ui/empty-state'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { usePageContext } from '@/features/layout/use-page-context'
import type { RowSelectionState } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Users } from 'lucide-react'
import { flattenDepartments, getDeptPath } from '@/lib/org'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export default function StructurePage() {
  const { open } = useWorkflow()
  const { setSubtitle } = usePageContext()
  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [departments, setDepartments] = useState<Department[]>([])
  const [members, setMembers] = useState<Member[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [directOnly, setDirectOnly] = useState(false)
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [approvalPendingCount, setApprovalPendingCount] = useState(0)
  const [confirmState, setConfirmState] = useState<{
    open: boolean
    title: string
    desc: string
    variant: 'primary' | 'danger'
    onConfirm: () => void
  }>({ open: false, title: '', desc: '', variant: 'primary', onConfirm: () => {} })
  const pageSize = 10

  const loadDepartments = async () => {
    const tree = await departmentApi.getTree()
    setDepartments(tree)
  }

  const loadMembers = useCallback(async () => {
    const params: Parameters<typeof memberApi.list>[0] = { page, pageSize }
    if (selectedDept) {
      params.departmentId = selectedDept.id
      params.directOnly = directOnly
    }
    const res: Paginated<Member> = await memberApi.list(params)
    setMembers(res.items)
    setTotal(res.total)
  }, [page, selectedDept, directOnly])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const tree = await departmentApi.getTree()
      if (!cancelled) setDepartments(tree)
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const params: Parameters<typeof memberApi.list>[0] = { page, pageSize }
      if (selectedDept) {
        params.departmentId = selectedDept.id
        params.directOnly = directOnly
      }
      const res = await memberApi.list(params)
      if (!cancelled) {
        setMembers(res.items)
        setTotal(res.total)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [page, selectedDept, directOnly])

  useEffect(() => {
    approvalApi.list({ tab: 'pending' }).then((items) => setApprovalPendingCount(items.length))
  }, [])

  useEffect(() => {
    if (selectedDept && departments.length > 0) {
      const path = getDeptPath(departments, selectedDept.id)
      setSubtitle(path)
    } else {
      setSubtitle(null)
    }
    return () => setSubtitle(null)
  }, [selectedDept, departments, setSubtitle])

  const handleSelectDept = (dept: Department | undefined) => {
    setSelectedDept(dept)
    setPage(1)
    setRowSelection({})
  }

  const inactiveCount = members.filter(
    (m) => m.status === 'inactive' || m.status === 'pending',
  ).length
  const selectedIds = Object.keys(rowSelection)

  const openMemberForm = (member?: Member | null) => {
    open('member-form', {
      member: member ?? null,
      departments,
      defaultDeptId: selectedDept?.id ?? '',
      onSubmit: async (data: {
        name: string
        phone: string
        email: string
        departmentId: string
      }) => {
        if (member) {
          await memberApi.update(member.id, data)
        } else {
          const dept = flattenDepartments(departments).find((d) => d.id === data.departmentId)
          await memberApi.create({ ...data, departmentName: dept?.name ?? '' })
        }
        loadMembers()
      },
    })
  }

  const handleStatusChange = (ids: string[], status: 'active' | 'inactive') => {
    const desc =
      status === 'inactive'
        ? `停用后该成员的 Platform Key 将同步失效`
        : `确定启用选中的 ${ids.length} 名成员？`
    setConfirmState({
      open: true,
      title: status === 'inactive' ? '停用成员' : '启用成员',
      desc,
      variant: status === 'inactive' ? 'danger' : 'primary',
      onConfirm: async () => {
        await memberApi.updateStatus(ids, status)
        setConfirmState((s) => ({ ...s, open: false }))
        setRowSelection({})
        loadMembers()
      },
    })
  }

  const handleDelete = (ids: string[]) => {
    setConfirmState({
      open: true,
      title: '删除成员',
      desc: `确定删除 ${ids.length} 名成员？此操作不可恢复`,
      variant: 'danger',
      onConfirm: async () => {
        await memberApi.delete(ids)
        setConfirmState((s) => ({ ...s, open: false }))
        setRowSelection({})
        loadMembers()
      },
    })
  }

  const handleInvite = async (inviteValue: string) => {
    const val = inviteValue.trim()
    if (!val) return
    const isEmail = val.includes('@')
    await memberApi.invite(isEmail ? { email: val } : { phone: val })
    loadMembers()
  }

  const handleBatchInvite = async () => {
    const inactiveIds = members
      .filter((m) => m.status === 'inactive' || m.status === 'pending')
      .map((m) => m.id)
    const result = await memberApi.batchInvite(inactiveIds.length > 0 ? inactiveIds : undefined)
    toast.success(`已向 ${result.sent} 名未激活成员发送邀请`)
  }

  const handleBatchTransfer = () => {
    open('pick-dept', {
      onConfirm: async (deptId: string) => {
        await memberApi.transferDepartment(selectedIds, deptId)
        setRowSelection({})
        loadMembers()
      },
    })
  }

  return (
    <div className="flex min-h-0 flex-1 gap-4">
      <DepartmentTree
        selectedId={selectedDept?.id}
        onSelect={handleSelectDept}
        onTreeChange={loadDepartments}
      />

      <div className="flex-1 flex flex-col gap-4 min-w-0">
        <Card size="sm">
          <CardContent className="flex items-center gap-6">
            <span className="text-base font-semibold">{selectedDept?.name ?? '全部成员'}</span>
            <span className="text-sm text-muted-foreground">
              总人数: <span className="font-medium text-foreground">{total}</span>
            </span>
            {inactiveCount > 0 && (
              <span className="text-sm text-yellow-600">未激活: {inactiveCount} 人</span>
            )}
            {approvalPendingCount > 0 && (
              <Link to="/keys/approval" className="text-sm text-indigo-600 hover:text-indigo-500">
                待审批申请: {approvalPendingCount} → 去审批
              </Link>
            )}
          </CardContent>
        </Card>

        {inactiveCount > 0 && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg px-4 py-2 text-sm text-yellow-800 flex items-center justify-between gap-3">
            <span>当前有 {inactiveCount} 名成员未激活</span>
            <Button size="sm" variant="outline" onClick={handleBatchInvite}>
              批量发送邀请
            </Button>
          </div>
        )}

        <div className="flex items-center gap-3 flex-wrap">
          <Select
            value={directOnly ? 'direct' : 'all'}
            onValueChange={(v) => {
              setDirectOnly(v === 'direct')
              setPage(1)
              setRowSelection({})
            }}
          >
            <SelectTrigger className="w-[100px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="direct">仅直属</SelectItem>
            </SelectContent>
          </Select>

          {selectedIds.length > 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="outline" size="sm" />}>
                批量操作 ({selectedIds.length})
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onClick={handleBatchTransfer}>批量转移部门</DropdownMenuItem>
                <DropdownMenuItem onClick={() => handleStatusChange(selectedIds, 'active')}>
                  批量启用
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => handleStatusChange(selectedIds, 'inactive')}>
                  批量停用
                </DropdownMenuItem>
                <DropdownMenuItem variant="destructive" onClick={() => handleDelete(selectedIds)}>
                  批量删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}

          <div className="flex-1" />
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              open('member-import', {
                defaultDeptName: selectedDept?.name,
                onSuccess: loadMembers,
              })
            }
          >
            导入成员
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              open('member-invite', {
                onSubmit: handleInvite,
              })
            }
          >
            邀请成员
          </Button>
          <Button size="sm" onClick={() => openMemberForm(null)}>
            添加成员
          </Button>
        </div>

        {members.length === 0 && total === 0 ? (
          <EmptyState
            icon={Users}
            title="暂无成员"
            description={
              selectedDept ? `${selectedDept.name} 下还没有成员` : '请先选择部门或添加成员'
            }
            actionLabel="添加成员"
            onAction={() => openMemberForm(null)}
          />
        ) : (
          <MemberTable
            data={members}
            total={total}
            page={page}
            pageSize={pageSize}
            onPageChange={setPage}
            onEdit={(m) => openMemberForm(m)}
            onStatusChange={handleStatusChange}
            onDelete={handleDelete}
            rowSelection={rowSelection}
            onRowSelectionChange={setRowSelection}
          />
        )}
      </div>

      <AlertDialog
        open={confirmState.open}
        onOpenChange={(isOpen) => {
          if (!isOpen) setConfirmState((s) => ({ ...s, open: false }))
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmState.title}</AlertDialogTitle>
            <AlertDialogDescription>{confirmState.desc}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setConfirmState((s) => ({ ...s, open: false }))}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmState.onConfirm}
              className={
                confirmState.variant === 'danger'
                  ? 'bg-destructive text-destructive-foreground hover:bg-destructive/80'
                  : ''
              }
            >
              确认
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
