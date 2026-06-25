import { useCallback, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { Member, Role } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoRole } from '@/features/demo'

export function useRolesPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { open } = useWorkflow()
  const { memberId, refreshSession } = useDemoRole()
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  const {
    data: initData,
    loading,
    error: rolesError,
    refresh: refreshInit,
    setData: setInitData,
  } = useAsyncResource(async () => {
    const [rolesData, permsData] = await Promise.all([
      apis.roleApi.list(),
      apis.roleApi.getPermissions(),
    ])
    return { roles: rolesData, permissions: permsData }
  }, [apis])

  const roles = initData?.roles ?? []
  const permissions = initData?.permissions ?? []
  const activeRoleId = selectedRoleId ?? roles[0]?.id ?? null

  const {
    data: members = [],
    loading: membersLoading,
    error: membersError,
    refresh: refreshMembers,
  } = useAsyncResource(async () => {
    if (!activeRoleId) return []
    return apis.roleApi.getMembers(activeRoleId)
  }, [apis, activeRoleId])

  const selectedRole = roles.find((r) => r.id === activeRoleId) ?? null
  const error = rolesError ?? membersError

  const refresh = useCallback(async () => {
    await Promise.all([refreshInit(), refreshMembers()])
  }, [refreshInit, refreshMembers])

  const refreshRoles = async () => {
    const updated = await apis.roleApi.list()
    setInitData((prev) => ({ roles: updated, permissions: prev?.permissions ?? [] }))
    return updated
  }

  const handleSelectRole = (role: Role) => {
    setSelectedRoleId(role.id)
  }

  const handleAddRole = () => {
    open('role-form', {
      role: null,
      permissions,
      onSubmit: async (data: { name: string; permissions: string[] }) => {
        await apis.roleApi.create(data)
        await refreshRoles()
      },
    })
  }

  const handleEditRole = (role: Role) => {
    open('role-form', {
      role,
      permissions,
      onSubmit: async (data: { name: string; permissions: string[] }) => {
        await apis.roleApi.update(role.id, data)
        await refreshRoles()
      },
    })
  }

  const handleDeleteRole = (role: Role) => {
    if (role.type === 'preset') return
    setDeleteConfirm(role)
  }

  const handleConfirmDelete = async () => {
    if (!deleteConfirm) return
    await apis.roleApi.delete(deleteConfirm.id)
    if (selectedRoleId === deleteConfirm.id) {
      setSelectedRoleId(null)
    }
    setDeleteConfirm(null)
    await refreshRoles()
    await refreshSession()
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
    void refreshMembers()
    await refreshRoles()
    if (removeConfirm.member.id === memberId) {
      await refreshSession()
    }
  }

  const handleAddMember = () => {
    if (!activeRoleId || !selectedRole) return
    open('role-add-member', {
      roleId: activeRoleId,
      roleName: selectedRole.name,
      existingMemberIds: members.map((m) => m.id),
      onSuccess: async () => {
        await refreshMembers()
        await refreshRoles()
        await refreshSession()
      },
    })
  }

  return {
    roles,
    activeRoleId,
    selectedRole,
    members,
    loading,
    membersLoading,
    error,
    refresh,
    deleteConfirm,
    removeConfirm,
    setDeleteConfirm,
    setRemoveConfirm,
    handleSelectRole,
    handleAddRole,
    handleEditRole,
    handleDeleteRole,
    handleConfirmDelete,
    handleRemoveMember,
    handleConfirmRemove,
    handleAddMember,
  }
}
