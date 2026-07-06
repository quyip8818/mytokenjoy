import { useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { Member, Permission, Role } from '@/api/types'

export function useRolesPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [members, setMembers] = useState<Member[]>([])
  const initializedRef = useRef(false)

  const [formOpen, setFormOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  const selectedRole = roles.find((r) => r.id === selectedRoleId) ?? null

  useEffect(() => {
    if (initializedRef.current) return
    initializedRef.current = true
    const init = async () => {
      const [rolesData, permsData] = await Promise.all([
        apis.roleApi.list(),
        apis.roleApi.getPermissions(),
      ])
      setRoles(rolesData)
      setPermissions(permsData)
      if (rolesData.length > 0) {
        setSelectedRoleId(rolesData[0].id)
        const membersData = await apis.roleApi.getMembers(rolesData[0].id)
        setMembers(membersData)
      }
    }
    void init()
  }, [apis])

  const handleSelectRole = (role: Role) => {
    setSelectedRoleId(role.id)
    void apis.roleApi.getMembers(role.id).then(setMembers)
  }

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
    const updated = await apis.roleApi.list()
    setRoles(updated)
  }

  const handleConfirmDelete = async () => {
    if (!deleteConfirm) return
    await apis.roleApi.delete(deleteConfirm.id)
    if (selectedRoleId === deleteConfirm.id) {
      setSelectedRoleId(null)
    }
    setDeleteConfirm(null)
    const updated = await apis.roleApi.list()
    setRoles(updated)
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
    if (selectedRoleId) {
      const membersData = await apis.roleApi.getMembers(selectedRoleId)
      setMembers(membersData)
    }
    const updated = await apis.roleApi.list()
    setRoles(updated)
  }

  const handleAddMember = async (memberId: string) => {
    if (!selectedRoleId) return
    await apis.roleApi.addMember(selectedRoleId, memberId)
    const membersData = await apis.roleApi.getMembers(selectedRoleId)
    setMembers(membersData)
    const updated = await apis.roleApi.list()
    setRoles(updated)
  }

  return {
    roles,
    permissions,
    selectedRoleId,
    selectedRole,
    members,
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
  }
}
