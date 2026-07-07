import { useCallback, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { Member, Role } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

export function useRolesPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [formOpen, setFormOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  const {
    data: roles = [],
    loading: rolesLoading,
    error: rolesError,
    refresh: refreshRoles,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.roles(),
    queryFn: (api) => api.roleApi.list(),
  })

  const {
    data: permissions = [],
    loading: permissionsLoading,
    error: permissionsError,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.permissions(),
    queryFn: (api) => api.roleApi.getPermissions(),
  })

  const resolvedSelectedRoleId = selectedRoleId ?? roles[0]?.id ?? null

  const {
    data: members = [],
    loading: membersLoading,
    error: membersError,
    refresh: refreshMembers,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.roleMembers(resolvedSelectedRoleId ?? ''),
    queryFn: (api) => api.roleApi.getMembers(resolvedSelectedRoleId!),
    enabled: Boolean(resolvedSelectedRoleId),
  })

  const selectedRole = roles.find((role) => role.id === resolvedSelectedRoleId) ?? null

  const invalidateOrg = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.org.all })
  }, [queryClient])

  const handleSelectRole = useCallback((role: Role) => {
    setSelectedRoleId(role.id)
  }, [])

  const handleAddRole = () => {
    setEditingRole(null)
    setFormOpen(true)
  }

  const handleEditRole = (role: Role) => {
    setEditingRole(role)
    setFormOpen(true)
  }

  const handleDeleteRole = (role: Role) => {
    if (role.type === 'preset') return
    setDeleteConfirm(role)
  }

  const handleFormSubmit = async (data: { name: string; permissions: string[] }) => {
    if (editingRole) {
      await apis.roleApi.update(editingRole.id, data)
    } else {
      await apis.roleApi.create(data)
    }
    setFormOpen(false)
    await invalidateOrg()
  }

  const handleConfirmDelete = async () => {
    if (!deleteConfirm) return
    await apis.roleApi.delete(deleteConfirm.id)
    if (resolvedSelectedRoleId === deleteConfirm.id) {
      setSelectedRoleId(null)
    }
    setDeleteConfirm(null)
    await invalidateOrg()
  }

  const handleRemoveMember = (member: Member) => {
    if (!selectedRole) return
    if (selectedRole.name === '普通成员') {
      toast('普通成员为保底角色，不可移除')
      return
    }
    setRemoveConfirm({ member, role: selectedRole })
  }

  const handleConfirmRemove = async () => {
    if (!removeConfirm) return
    await apis.roleApi.removeMember(removeConfirm.role.id, removeConfirm.member.id)
    setRemoveConfirm(null)
    await invalidateOrg()
  }

  const handleAddMember = async (memberId: string) => {
    if (!resolvedSelectedRoleId) return
    await apis.roleApi.addMember(resolvedSelectedRoleId, memberId)
    await invalidateOrg()
  }

  const searchMembers = async (keyword: string) => {
    const res = await apis.memberApi.list({ page: 1, pageSize: 20, keyword })
    return res.items
  }

  const refresh = useCallback(async () => {
    await Promise.all([refreshRoles(), refreshMembers()])
  }, [refreshRoles, refreshMembers])

  return {
    roles,
    permissions,
    selectedRoleId: resolvedSelectedRoleId,
    selectedRole,
    members,
    loading: rolesLoading || permissionsLoading || membersLoading,
    error: rolesError ?? permissionsError ?? membersError,
    rolesLoading,
    rolesError,
    membersLoading,
    membersError,
    formOpen,
    editingRole,
    deleteConfirm,
    addMemberOpen,
    removeConfirm,
    setFormOpen,
    setDeleteConfirm,
    setAddMemberOpen,
    setRemoveConfirm,
    handleSelectRole,
    handleAddRole,
    handleEditRole,
    handleDeleteRole,
    handleFormSubmit,
    handleConfirmDelete,
    handleRemoveMember,
    handleConfirmRemove,
    handleAddMember,
    searchMembers,
    refresh,
    refreshRoles,
    refreshMembers,
  }
}
