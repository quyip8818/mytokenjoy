import { useEffect, useState } from 'react'
import { useOrgStructureStore, initOrgStructureApis } from '@/stores/org-structure'
import type { Department, Member } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'

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

export function useStructurePage() {
  const apis = useInjectedApis()
  const store = useOrgStructureStore()

  const [formOpen, setFormOpen] = useState(false)
  const [editingMember, setEditingMember] = useState<Member | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferDeptId, setTransferDeptId] = useState('')
  const [confirmState, setConfirmState] = useState<ConfirmState>(INITIAL_CONFIRM_STATE)

  useEffect(() => {
    initOrgStructureApis(apis)
    void store.loadDepartments()
    void store.loadMembers()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apis])

  const pendingCount = store.members.filter((m) => m.status === 'pending').length
  const selectedIds = Object.keys(store.rowSelection)
  const flatDepts = flattenDepts(store.departments)

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
      await store.updateMember(editingMember.id, {
        name: data.name,
        phone: data.phone,
        email: data.email,
        departmentId: data.departmentId,
        departmentName:
          flattenDepts(store.departments).find((d) => d.id === data.departmentId)?.name ?? '',
      })
    } else {
      const dept = flattenDepts(store.departments).find((d) => d.id === data.departmentId)
      await store.createMember({
        companyId: 1,
        name: data.name,
        phone: data.phone,
        email: data.email,
        departmentId: data.departmentId,
        departmentName: dept?.name ?? '',
      })
    }
    setFormOpen(false)
    setEditingMember(null)
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
        await store.updateMemberStatus(ids, status)
        setConfirmState((s) => ({ ...s, open: false }))
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
        await store.deleteMember(ids)
        setConfirmState((s) => ({ ...s, open: false }))
      },
    })
  }

  const handleBatchTransfer = async () => {
    if (!transferDeptId || selectedIds.length === 0) return
    await store.transferMembers(selectedIds, transferDeptId)
    setTransferOpen(false)
    setTransferDeptId('')
  }

  const openCreateMember = () => {
    setEditingMember(null)
    setFormOpen(true)
  }

  const openEditMember = (member: Member) => {
    setEditingMember(member)
    setFormOpen(true)
  }

  const closeMemberForm = () => {
    setFormOpen(false)
    setEditingMember(null)
  }

  return {
    store,
    formOpen,
    editingMember,
    inviteOpen,
    transferOpen,
    transferDeptId,
    confirmState,
    pendingCount,
    selectedIds,
    flatDepts,
    setInviteOpen,
    setTransferOpen,
    setTransferDeptId,
    setConfirmState,
    handleMemberSubmit,
    handleStatusChange,
    handleDelete,
    handleBatchTransfer,
    openCreateMember,
    openEditMember,
    closeMemberForm,
  }
}
