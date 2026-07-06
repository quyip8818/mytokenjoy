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
import { Shield } from 'lucide-react'
import { useRolesPage } from './hooks/use-roles-page'

export default function RolesPage() {
  const {
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
  } = useRolesPage()

  return (
    <div className="flex h-full rounded-lg border border-border bg-card shadow-xs overflow-hidden">
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
            onAddMember={() => setAddMemberOpen(true)}
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-2">
            <Shield className="h-10 w-10 text-muted-foreground/30" strokeWidth={1.5} />
            <p className="text-sm text-muted-foreground">请选择一个角色</p>
          </div>
        )}
      </div>

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
