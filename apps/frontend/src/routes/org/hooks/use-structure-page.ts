import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import type { RowSelectionState } from '@tanstack/react-table'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { Department, Member } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { usePageSubtitle } from '@/hooks/use-page-subtitle'
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

export function useStructurePage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { open } = useWorkflow()
  const { setSubtitle } = usePageSubtitle()
  const { canWrite, permissions } = usePermissions()
  const canApprove = permissions.includes(PERMISSION.BUDGET_APPROVE)
  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [page, setPage] = useState(1)
  const [directOnly, setDirectOnly] = useState(false)
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [confirmState, setConfirmState] = useState<ConfirmActionState>(INITIAL_CONFIRM_STATE)

  const {
    data: departments = [],
    error: deptError,
    refresh: refreshDepartments,
  } = useAsyncResource(() => apis.departmentApi.getTree(), [apis])

  const {
    data: memberData,
    loading: membersLoading,
    error: memberError,
    refresh: refreshMembers,
  } = useAsyncResource(async () => {
    const params: Parameters<typeof apis.memberApi.list>[0] = { page, pageSize: PAGE_SIZE }
    if (selectedDept) {
      params.departmentId = selectedDept.id
      params.directOnly = directOnly
    }
    return apis.memberApi.list(params)
  }, [apis, page, selectedDept, directOnly])

  const members = memberData?.items ?? []
  const total = memberData?.total ?? 0

  const { data: approvalPendingCount = 0 } = useAsyncResource(async () => {
    if (!canApprove) return 0
    const items = await apis.approvalApi.list({ tab: 'pending' })
    return items.length
  }, [apis, canApprove])

  const error = memberError ?? deptError

  const refresh = useCallback(async () => {
    await Promise.all([refreshDepartments(), refreshMembers()])
  }, [refreshDepartments, refreshMembers])

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
          await apis.memberApi.update(member.id, data)
        } else {
          const dept = flattenDepartments(departments).find((d) => d.id === data.departmentId)
          await apis.memberApi.create({ ...data, departmentName: dept?.name ?? '' })
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
        await apis.memberApi.updateStatus(ids, status)
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
        await apis.memberApi.delete(ids)
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
    await apis.memberApi.invite(isEmail ? { email: val } : { phone: val })
    refreshMembers()
  }

  const handleBatchInvite = async () => {
    const inactiveIds = members
      .filter((m) => m.status === 'inactive' || m.status === 'pending')
      .map((m) => m.id)
    const result = await apis.memberApi.batchInvite(
      inactiveIds.length > 0 ? inactiveIds : undefined,
    )
    toast.success(`已向 ${result.sent} 名未激活成员发送邀请`)
  }

  const handleBatchTransfer = () => {
    open('pick-dept', {
      onConfirm: async (deptId: string) => {
        await apis.memberApi.transferDepartment(selectedIds, deptId)
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
    error,
    refresh,
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
