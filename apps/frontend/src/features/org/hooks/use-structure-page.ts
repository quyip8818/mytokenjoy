import { useCallback, useMemo, useState } from 'react'
import type { RowSelectionState } from '@tanstack/react-table'
import { useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { ApiError } from '@/api/client'
import type { Department, Member } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { flattenDepartments } from '../lib/departments'
import { useStructureConfirmState } from './use-structure-confirm'

const DEFAULT_PAGE_SIZE = 10

export function useStructurePage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const [selectedDept, setSelectedDept] = useState<Department | undefined>()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)
  const [keyword, setKeyword] = useState('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [formOpen, setFormOpen] = useState(false)
  const [editingMember, setEditingMember] = useState<Member | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferDeptId, setTransferDeptId] = useState('')
  const { confirmState, setConfirmState } = useStructureConfirmState()

  const memberQueryParams = useMemo(
    () => ({
      page,
      pageSize,
      keyword: searchKeyword || undefined,
      departmentId: selectedDept?.id,
    }),
    [page, pageSize, searchKeyword, selectedDept?.id],
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
    placeholderData: keepPreviousData,
  })

  const members = membersResult?.items ?? []
  const total = membersResult?.total ?? 0

  const invalidateOrg = useCallback(async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.org.departmentTree() }),
      queryClient.invalidateQueries({ queryKey: queryKeys.org.members(memberQueryParams) }),
    ])
  }, [queryClient, memberQueryParams])

  const selectDept = useCallback((dept: Department | undefined) => {
    setSelectedDept(dept)
    setPage(1)
    setRowSelection({})
  }, [])

  const setKeywordAndReset = useCallback((value: string) => {
    setKeyword(value)
  }, [])

  const handleSearch = useCallback(() => {
    setSearchKeyword(keyword)
    setPage(1)
    setRowSelection({})
  }, [keyword])

  const setPageAndRefresh = useCallback((nextPage: number) => {
    setPage(nextPage)
  }, [])

  const pendingCount = membersResult?.pendingCount ?? 0
  const selectedIds = Object.keys(rowSelection)
  const flatDepts = flattenDepartments(departments)

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
    try {
      if (editingMember) {
        await apis.memberApi.update(editingMember.id, {
          alias: data.name,
          phone: data.phone,
          email: data.email,
          departmentId: data.departmentId,
          departmentName:
            flattenDepartments(departments).find((dept) => dept.id === data.departmentId)?.name ??
            '',
        })
        toast.success(`成员「${data.name}」已更新`)
      } else {
        const dept = flattenDepartments(departments).find((item) => item.id === data.departmentId)
        await apis.memberApi.create({
          alias: data.name,
          phone: data.phone,
          email: data.email,
          departmentId: data.departmentId,
          departmentName: dept?.name ?? '',
        })
        toast.success(`成员「${data.name}」添加成功`)
      }
      setFormOpen(false)
      setEditingMember(null)
      await invalidateOrg()
    } catch (err) {
      const message = err instanceof ApiError ? err.message : '操作失败，请重试'
      toast.error(message)
    }
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
      desc: `确定删除 ${ids.length} 名成员？删除后不可恢复`,
      variant: 'danger',
      onConfirm: async () => {
        try {
          await apis.memberApi.delete(ids)
          setRowSelection({})
          setConfirmState((state) => ({ ...state, open: false }))
          toast.success(`已删除 ${ids.length} 名成员`)
          // If current page would be empty after deletion, go back one page
          const remainingTotal = total - ids.length
          const maxPage = Math.max(1, Math.ceil(remainingTotal / pageSize))
          if (page > maxPage) {
            setPage(maxPage)
          }
          await invalidateOrg()
        } catch (err) {
          const message = err instanceof ApiError ? err.message : '删除失败，请重试'
          toast.error(message)
        }
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
      try {
        await apis.departmentApi.create({ name, parentId })
        toast.success(`部门「${name}」创建成功`)
        await invalidateOrg()
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '创建部门失败'
        toast.error(message)
      }
    },
    [apis, invalidateOrg],
  )

  const updateDept = useCallback(
    async (id: string, name: string) => {
      try {
        await apis.departmentApi.update(id, { name })
        toast.success(`部门已更新为「${name}」`)
        await invalidateOrg()
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '更新部门失败'
        toast.error(message)
      }
    },
    [apis, invalidateOrg],
  )

  const deleteDept = useCallback(
    async (id: string) => {
      try {
        await apis.departmentApi.delete(id)
        toast.success('部门已删除')
        if (selectedDept?.id === id) {
          setSelectedDept(undefined)
        }
        await invalidateOrg()
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '删除部门失败'
        toast.error(message)
      }
    },
    [apis, invalidateOrg, selectedDept],
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
    pageSize,
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
    handleSearch,
    setPage: setPageAndRefresh,
    setPageSize: (size: number) => {
      setPageSize(size)
      setPage(1)
    },
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
