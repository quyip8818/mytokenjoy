import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { toast } from 'sonner'
import type { Department, Member } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { departmentApi, memberApi } from '@/api/org'
import { approvalApi } from '@/api/keys'
import { DepartmentTree } from '@/components/org/department-tree'
import { MemberTable } from '@/components/org/member-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { usePageSubtitle } from '@/lib/page-subtitle'
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
import { PermissionGate } from '@/components/auth/permission-gate'
import { usePermissions } from '@/hooks/use-permissions'
import { PERMISSION } from '@/lib/permissions'

export default function StructurePage() {
  const { open } = useWorkflow()
  const { setSubtitle } = usePageSubtitle()
  const { canWrite, permissions } = usePermissions()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)
  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [page, setPage] = useState(1)
  const [directOnly, setDirectOnly] = useState(false)
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [confirmState, setConfirmState] = useState<{
    open: boolean
    title: string
    desc: string
    variant: 'primary' | 'danger'
    onConfirm: () => void
  }>({ open: false, title: '', desc: '', variant: 'primary', onConfirm: () => {} })
  const pageSize = 10

  const { data: departments = [], refresh: refreshDepartments } = useAsyncResource(
    () => departmentApi.getTree(),
    [],
  )

  const {
    data: memberData,
    loading: membersLoading,
    refresh: refreshMembers,
  } = useAsyncResource(async () => {
    const params: Parameters<typeof memberApi.list>[0] = { page, pageSize }
    if (selectedDept) {
      params.departmentId = selectedDept.id
      params.directOnly = directOnly
    }
    return memberApi.list(params)
  }, [page, selectedDept, directOnly])

  const members = memberData?.items ?? []
  const total = memberData?.total ?? 0

  const { data: approvalPendingCount = 0 } = useAsyncResource(async () => {
    if (!canApprove) return 0
    const items = await approvalApi.list({ tab: 'pending' })
    return items.length
  }, [canApprove])

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
        refreshMembers()
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
        refreshMembers()
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
        refreshMembers()
      },
    })
  }

  const handleInvite = async (inviteValue: string) => {
    const val = inviteValue.trim()
    if (!val) return
    const isEmail = val.includes('@')
    await memberApi.invite(isEmail ? { email: val } : { phone: val })
    refreshMembers()
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
        refreshMembers()
      },
    })
  }

  const handlePageChange = (nextPage: number) => {
    setPage(nextPage)
  }

  return (
    <PageShell
      layout="split"
      sidebar={
        <DepartmentTree
          selectedId={selectedDept?.id}
          onSelect={handleSelectDept}
          onTreeChange={refreshDepartments}
          readOnly={!canWrite}
        />
      }
    >
      <Card size="sm" className="border-border/50 shadow-card">
        <CardContent className="flex items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-6">
            <span className="text-base font-semibold">{selectedDept?.name ?? '全部成员'}</span>
            <span className="text-sm text-muted-foreground">
              总人数: <span className="font-medium text-foreground">{total}</span>
            </span>
            {inactiveCount > 0 && (
              <StatusBadge variant="warning">未激活 {inactiveCount} 人</StatusBadge>
            )}
            {approvalPendingCount > 0 && (
              <PermissionGate permission={PERMISSION.BUDGET_APPROVE}>
                <Link to="/keys/approval" className="text-sm text-blue-600 hover:text-blue-500">
                  待审批申请: {approvalPendingCount} → 去审批
                </Link>
              </PermissionGate>
            )}
          </div>
          <PermissionGate write>
            {inactiveCount > 0 && (
              <Button size="sm" variant="outline" onClick={handleBatchInvite}>
                批量发送邀请
              </Button>
            )}
          </PermissionGate>
        </CardContent>
      </Card>

      <div className="flex flex-wrap items-center gap-3">
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
          <PermissionGate write>
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
          </PermissionGate>
        )}

        <div className="flex-1" />
        <PermissionGate write>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              open('member-import', {
                defaultDeptName: selectedDept?.name,
                onSuccess: refreshMembers,
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
          <Button size="sm" variant="brand" onClick={() => openMemberForm(null)}>
            添加成员
          </Button>
        </PermissionGate>
      </div>

      <DataSection
        loading={membersLoading}
        skeletonColumns={6}
        empty={
          !membersLoading && members.length === 0 && total === 0
            ? {
                icon: Users,
                title: '暂无成员',
                description: selectedDept
                  ? `${selectedDept.name} 下还没有成员`
                  : '请先选择部门或添加成员',
                actionLabel: canWrite ? '添加成员' : undefined,
                onAction: canWrite ? () => openMemberForm(null) : undefined,
              }
            : null
        }
      >
        <MemberTable
          data={members}
          total={total}
          page={page}
          pageSize={pageSize}
          onPageChange={handlePageChange}
          onEdit={(m) => openMemberForm(m)}
          onStatusChange={handleStatusChange}
          onDelete={handleDelete}
          rowSelection={rowSelection}
          onRowSelectionChange={setRowSelection}
          readOnly={!canWrite}
        />
      </DataSection>

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
    </PageShell>
  )
}
