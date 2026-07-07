import { useCallback, useMemo, useState } from 'react'
import type { RowSelectionState } from '@tanstack/react-table'
import { useQueryClient } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import type { Department, Member } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'

type ConfirmState = {
  open: boolean
  title: string
  desc: string
  variant: 'primary' | 'danger'
  onConfirm: () => void
}

const INITIAL_CONFIRM_STATE: ConfirmState = {
  open: false,
  title: '',
  desc: '',
  variant: 'primary',
  onConfirm: () => {},
}

const PAGE_SIZE = 10

function flattenDepts(
  departments: Department[],
  level = 0,
): { id: string; name: string; level: number }[] {
  const result: { id: string; name: string; level: number }[] = []
  for (const dept of departments) {
    result.push({ id: dept.id, name: dept.name, level })
    if (dept.children) result.push(...flattenDepts(dept.children, level + 1))
  }
  return result
}

export function useStructurePage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [formOpen, setFormOpen] = useState(false)
  const [editingMember, setEditingMember] = useState<Member | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferDeptId, setTransferDeptId] = useState('')
  const [confirmState, setConfirmState] = useState<ConfirmState>(INITIAL_CONFIRM_STATE)

  const memberQueryParams = useMemo(
    () => ({
      page,
      pageSize: PAGE_SIZE,
      keyword: keyword || undefined,
      departmentId: selectedDept?.id,
    }),
    [page, keyword, selectedDept?.id],
  )

  const {
    data: departments = [],
    loading: departmentsLoading,
    error: departmentsError,
    refresh: refreshDepartments,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.departmentTree(),
    queryFn: (api) => api.departmentApi.getTree(),
  })

  const {
    data: membersResult,
    loading: membersLoading,
    error: membersError,
    refresh: refreshMembers,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.members(memberQueryParams),
    queryFn: (api) => api.memberApi.list(memberQueryParams),
  })

  const members = membersResult?.items ?? []
  const total = membersResult?.total ?? 0

  const invalidateOrg = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.org.all })
  }, [queryClient])

  const selectDept = useCallback((dept: Department | undefined) => {
    setSelectedDept(dept)
    setPage(1)
    setRowSelection({})
  }, [])

  const setKeywordAndReset = useCallback((value: string) => {
    setKeyword(value)
    setPage(1)
    setRowSelection({})
  }, [])

  const setPageAndRefresh = useCallback((nextPage: number) => {
    setPage(nextPage)
  }, [])

  const pendingCount = members.filter((member) => member.status === 'pending').length
  const selectedIds = Object.keys(rowSelection)
  const flatDepts = flattenDepts(departments)

  const handleMemberSubmit = async (data: {
    name: string
    phone: string
    email: string
    username: string
    employeeId: string
    jobTitle: string
    hireDate: string
    departmentId: string
  }) => {
    if (editingMember) {
      await apis.memberApi.update(editingMember.id, {
        name: data.name,
        phone: data.phone,
        email: data.email,
        departmentId: data.departmentId,
        departmentName:
          flattenDepts(departments).find((dept) => dept.id === data.departmentId)?.name ?? '',
      })
    } else {
      const dept = flattenDepts(departments).find((item) => item.id === data.departmentId)
      await apis.memberApi.create({
        name: data.name,
        phone: data.phone,
        email: data.email,
        departmentId: data.departmentId,
        departmentName: dept?.name ?? '',
      })
    }
    setFormOpen(false)
    setEditingMember(null)
    await invalidateOrg()
  }

  const handleStatusChange = (ids: string[], status: 'active' | 'inactive') => {
    setConfirmState({
      open: true,
      title: status === 'inactive' ? '停用成员' : '启用成员',
      desc:
        status === 'inactive'
          ? '停用后该成员的 Platform Key 将同步失效'
          : `确定启用选中的 ${ids.length} 名成员？`,
      variant: status === 'inactive' ? 'danger' : 'primary',
      onConfirm: async () => {
        await apis.memberApi.updateStatus(ids, status)
        setRowSelection({})
        setConfirmState((state) => ({ ...state, open: false }))
        await invalidateOrg()
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
        setRowSelection({})
        setConfirmState((state) => ({ ...state, open: false }))
        await invalidateOrg()
      },
    })
  }

  const handleBatchTransfer = async () => {
    if (!transferDeptId || selectedIds.length === 0) return
    await apis.memberApi.transferDepartment(selectedIds, transferDeptId)
    setTransferOpen(false)
    setTransferDeptId('')
    setRowSelection({})
    await invalidateOrg()
  }

  const inviteMember = async (value: string) => {
    const isEmail = value.includes('@')
    await apis.memberApi.invite(isEmail ? { email: value } : { phone: value })
    await invalidateOrg()
  }

  const openCreateMember = () => {
    setEditingMember(null)
    setFormOpen(true)
  }

  const openEditMember = (member: Member) => {
    setEditingMember(member)
    setFormOpen(true)
  }

  const createDept = useCallback(
    async (name: string, parentId: string) => {
      await apis.departmentApi.create({ name, parentId })
      await invalidateOrg()
    },
    [apis, invalidateOrg],
  )

  const updateDept = useCallback(
    async (id: string, name: string) => {
      await apis.departmentApi.update(id, { name })
      await invalidateOrg()
    },
    [apis, invalidateOrg],
  )

  const deleteDept = useCallback(
    async (id: string) => {
      await apis.departmentApi.delete(id)
      if (selectedDept?.id === id) {
        setSelectedDept(undefined)
      }
      await invalidateOrg()
    },
    [apis, invalidateOrg, selectedDept?.id],
  )

  const closeMemberForm = () => {
    setFormOpen(false)
    setEditingMember(null)
  }

  const loading = departmentsLoading || membersLoading
  const error = departmentsError ?? membersError
  const refresh = useCallback(async () => {
    await Promise.all([refreshDepartments(), refreshMembers()])
  }, [refreshDepartments, refreshMembers])

  return {
    departments,
    selectedDept,
    members,
    total,
    page,
    pageSize: PAGE_SIZE,
    keyword,
    rowSelection,
    loading,
    error,
    departmentsLoading,
    departmentsError,
    membersLoading,
    membersError,
    formOpen,
    editingMember,
    inviteOpen,
    transferOpen,
    transferDeptId,
    confirmState,
    pendingCount,
    selectedIds,
    flatDepts,
    selectDept,
    createDept,
    updateDept,
    deleteDept,
    setKeyword: setKeywordAndReset,
    setPage: setPageAndRefresh,
    setRowSelection,
    refresh,
    refreshDepartments,
    refreshMembers,
    setInviteOpen,
    setTransferOpen,
    setTransferDeptId,
    setConfirmState,
    handleMemberSubmit,
    handleStatusChange,
    handleDelete,
    handleBatchTransfer,
    inviteMember,
    openCreateMember,
    openEditMember,
    closeMemberForm,
  }
}
