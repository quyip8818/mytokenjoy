import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import type { RowSelectionState } from '@tanstack/react-table'
import type { Department, Member } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { departmentApi, memberApi } from '@/api/org'
import { approvalApi } from '@/api/keys'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { usePageSubtitle } from '@/lib/page-subtitle'
import { flattenDepartments, getDeptPath } from '@/lib/org'
import { usePermissions } from '@/hooks/use-permissions'
import { PERMISSION } from '@/lib/permissions'
import type { ConfirmActionState } from '@/components/ui/confirm-action-dialog'

const PAGE_SIZE = 10

const INITIAL_CONFIRM_STATE: ConfirmActionState = {
  open: false,
  title: '',
  desc: '',
  variant: 'primary',
  onConfirm: () => {},
}

export function useStructurePage() {
  const { open } = useWorkflow()
  const { setSubtitle } = usePageSubtitle()
  const { canWrite, permissions } = usePermissions()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)
  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [page, setPage] = useState(1)
  const [directOnly, setDirectOnly] = useState(false)
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [confirmState, setConfirmState] = useState<ConfirmActionState>(INITIAL_CONFIRM_STATE)

  const { data: departments = [], refresh: refreshDepartments } = useAsyncResource(
    () => departmentApi.getTree(),
    [],
  )

  const {
    data: memberData,
    loading: membersLoading,
    refresh: refreshMembers,
  } = useAsyncResource(async () => {
    const params: Parameters<typeof memberApi.list>[0] = { page, pageSize: PAGE_SIZE }
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

  const setDirectOnlyFilter = (nextDirectOnly: boolean) => {
    setDirectOnly(nextDirectOnly)
    setPage(1)
    setRowSelection({})
  }

  const closeConfirm = () => {
    setConfirmState((s) => ({ ...s, open: false }))
  }

  return {
    canWrite,
    selectedDept,
    departments,
    members,
    total,
    page,
    pageSize: PAGE_SIZE,
    membersLoading,
    directOnly,
    rowSelection,
    confirmState,
    inactiveCount,
    selectedIds,
    approvalPendingCount,
    refreshDepartments,
    refreshMembers,
    handleSelectDept,
    openMemberForm,
    handleStatusChange,
    handleDelete,
    handleBatchInvite,
    handleBatchTransfer,
    handleInvite,
    setDirectOnlyFilter,
    setPage,
    setRowSelection,
    closeConfirm,
    open,
  }
}
