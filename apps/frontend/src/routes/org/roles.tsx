import { useEffect, useRef, useState } from 'react'
import type { Member, Permission, Role } from '@/api/types'
import { roleApi } from '@/api/org'
import { RoleList } from '@/components/org/role-list'
import { RoleForm } from '@/components/org/role-form'
import { RoleMemberTable, AddMemberDialog } from '@/components/org/role-member-table'
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
import { Shield } from 'lucide-react'

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [members, setMembers] = useState<Member[]>([])
  const initializedRef = useRef(false)

  // Dialog states
  const [formOpen, setFormOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  // Derive selectedRole from roles + selectedRoleId
  const selectedRole = roles.find((r) => r.id === selectedRoleId) ?? null

  // Initial data load
  useEffect(() => {
    if (initializedRef.current) return
    initializedRef.current = true
    const init = async () => {
      const [rolesData, permsData] = await Promise.all([
        roleApi.list(),
        roleApi.getPermissions(),
      ])
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

  // Handlers
  const handleSelectRole = (role: Role) => {
    setSelectedRoleId(role.id)
    roleApi.getMembers(role.id).then(setMembers)
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
      await roleApi.update(editingRole.id, data)
    } else {
      await roleApi.create(data)
    }
    setFormOpen(false)
    const updated = await roleApi.list()
    setRoles(updated)
  }

  const handleConfirmDelete = async () => {
    if (!deleteConfirm) return
    await roleApi.delete(deleteConfirm.id)
    if (selectedRoleId === deleteConfirm.id) {
      setSelectedRoleId(null)
    }
    setDeleteConfirm(null)
    const updated = await roleApi.list()
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
    await roleApi.removeMember(removeConfirm.role.id, removeConfirm.member.id)
    setRemoveConfirm(null)
    if (selectedRoleId) {
      const membersData = await roleApi.getMembers(selectedRoleId)
      setMembers(membersData)
    }
    const updated = await roleApi.list()
    setRoles(updated)
  }

  const handleAddMember = async (memberId: string) => {
    if (!selectedRoleId) return
    await roleApi.addMember(selectedRoleId, memberId)
    const membersData = await roleApi.getMembers(selectedRoleId)
    setMembers(membersData)
    const updated = await roleApi.list()
    setRoles(updated)
  }

  return (
    <div className="flex h-full rounded-lg border border-border bg-card shadow-xs overflow-hidden">
      {/* Left: Role list */}
      <RoleList
        roles={roles}
        selectedRoleId={selectedRoleId}
        onSelect={handleSelectRole}
        onAdd={handleAddRole}
        onEdit={handleEditRole}
        onDelete={handleDeleteRole}
      />

      {/* Right: Role detail */}
      <div className="flex-1 p-6 overflow-auto">
        {selectedRole ? (
          <RoleMemberTable
            role={selectedRole}
            members={members}
            onRemoveMember={handleRemoveMember}
            onAddMember={() => setAddMemberOpen(true)}
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-2">
            <Shield className="h-10 w-10 text-muted-foreground/30" strokeWidth={1.5} />
            <p className="text-sm text-muted-foreground">请选择一个角色</p>
          </div>
        )}
      </div>

      {/* Dialogs */}
      <RoleForm
        open={formOpen}
        role={editingRole}
        permissions={permissions}
        onSubmit={handleFormSubmit}
        onCancel={() => setFormOpen(false)}
      />

      {selectedRoleId && (
        <AddMemberDialog
          open={addMemberOpen}
          roleId={selectedRoleId}
          existingMemberIds={members.map((m) => m.id)}
          onAdd={handleAddMember}
          onClose={() => setAddMemberOpen(false)}
        />
      )}

      {/* Delete role confirmation */}
      <AlertDialog open={!!deleteConfirm} onOpenChange={(o) => { if (!o) setDeleteConfirm(null) }}>
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

      {/* Remove member confirmation */}
      <AlertDialog open={!!removeConfirm} onOpenChange={(o) => { if (!o) setRemoveConfirm(null) }}>
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
