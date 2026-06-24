import { useEffect, useRef, useState } from 'react'
import type { Member, Permission, Role } from '@/api/types'
import { roleApi } from '@/api/org'
import { RoleList } from '@/components/org/role-list'
import { RoleMemberTable } from '@/components/org/role-member-table'
import { EmptyState } from '@/components/ui/empty-state'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { Shield } from 'lucide-react'
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
import { toast } from 'sonner'

export default function RolesPage() {
  const { open } = useWorkflow()
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [members, setMembers] = useState<Member[]>([])
  const initializedRef = useRef(false)

  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  const selectedRole = roles.find((r) => r.id === selectedRoleId) ?? null

  const refreshRoles = async () => {
    const updated = await roleApi.list()
    setRoles(updated)
    return updated
  }

  useEffect(() => {
    if (initializedRef.current) return
    initializedRef.current = true
    const init = async () => {
      const [rolesData, permsData] = await Promise.all([roleApi.list(), roleApi.getPermissions()])
      setRoles(rolesData)
      setPermissions(permsData)
      if (rolesData.length > 0) {
        setSelectedRoleId(rolesData[0].id)
        const membersData = await roleApi.getMembers(rolesData[0].id)
        setMembers(membersData)
      }
    }
    init()
  }, [])

  const handleSelectRole = (role: Role) => {
    setSelectedRoleId(role.id)
    roleApi.getMembers(role.id).then(setMembers)
  }

  const handleAddRole = () => {
    open('role-form', {
      role: null,
      permissions,
      onSubmit: async (data: { name: string; permissions: string[] }) => {
        await roleApi.create(data)
        await refreshRoles()
      },
    })
  }

  const handleEditRole = (role: Role) => {
    open('role-form', {
      role,
      permissions,
      onSubmit: async (data: { name: string; permissions: string[] }) => {
        await roleApi.update(role.id, data)
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
    await roleApi.delete(deleteConfirm.id)
    if (selectedRoleId === deleteConfirm.id) {
      setSelectedRoleId(null)
      setMembers([])
    }
    setDeleteConfirm(null)
    await refreshRoles()
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
    await roleApi.removeMember(removeConfirm.role.id, removeConfirm.member.id)
    setRemoveConfirm(null)
    if (selectedRoleId) {
      const membersData = await roleApi.getMembers(selectedRoleId)
      setMembers(membersData)
    }
    await refreshRoles()
  }

  const handleAddMember = () => {
    if (!selectedRoleId || !selectedRole) return
    open('role-add-member', {
      roleId: selectedRoleId,
      roleName: selectedRole.name,
      existingMemberIds: members.map((m) => m.id),
      onSuccess: async () => {
        const membersData = await roleApi.getMembers(selectedRoleId)
        setMembers(membersData)
        await refreshRoles()
      },
    })
  }

  return (
    <div className="flex h-full rounded-lg border border-border overflow-hidden">
      <RoleList
        roles={roles}
        selectedRoleId={selectedRoleId}
        onSelect={handleSelectRole}
        onAdd={handleAddRole}
        onEdit={handleEditRole}
        onDelete={handleDeleteRole}
      />

      <div className="flex-1 p-6 overflow-auto">
        {selectedRole ? (
          <RoleMemberTable
            role={selectedRole}
            members={members}
            onRemoveMember={handleRemoveMember}
            onAddMember={handleAddMember}
          />
        ) : (
          <EmptyState
            icon={Shield}
            title="请选择一个角色"
            description="从左侧列表选择角色查看成员，或创建新角色"
            actionLabel="创建角色"
            onAction={handleAddRole}
          />
        )}
      </div>

      <AlertDialog
        open={!!deleteConfirm}
        onOpenChange={(o) => {
          if (!o) setDeleteConfirm(null)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除角色</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteConfirm && deleteConfirm.memberCount > 0
                ? `该角色下有 ${deleteConfirm.memberCount} 名成员，删除后将失去对应权限，是否继续？`
                : '确定要删除该角色吗？'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleConfirmDelete}>
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog
        open={!!removeConfirm}
        onOpenChange={(o) => {
          if (!o) setRemoveConfirm(null)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>移除成员</AlertDialogTitle>
            <AlertDialogDescription>
              {`确定将「${removeConfirm?.member.name ?? ''}」从「${removeConfirm?.role.name ?? ''}」角色中移除吗？`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleConfirmRemove}>
              移除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
