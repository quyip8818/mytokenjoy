import { RoleList, RoleForm, RoleMemberTable, AddMemberDialog, useRolesPage } from '@/features/org'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
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

export default function RolesPage() {
  const {
    roles,
    permissions,
    selectedRoleId,
    selectedRole,
    members,
    rolesLoading,
    rolesError,
    membersLoading,
    membersError,
    refreshRoles,
    refreshMembers,
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
  } = useRolesPage()

  return (
    <PageShell layout="fill">
      <div className="flex min-h-0 flex-1 overflow-hidden rounded-lg border border-border bg-card shadow-xs">
        <DataSection
          loading={rolesLoading}
          error={rolesError}
          onRetry={() => void refreshRoles()}
          className="shrink-0 rounded-none border-0 shadow-none"
          contentClassName="h-full p-0"
          loadingVariant="spinner"
        >
          <RoleList
            roles={roles}
            selectedRoleId={selectedRoleId}
            onSelect={handleSelectRole}
            onAdd={handleAddRole}
            onEdit={handleEditRole}
            onDelete={handleDeleteRole}
          />
        </DataSection>

        <DataSection
          loading={membersLoading}
          error={membersError}
          onRetry={() => void refreshMembers()}
          className="flex min-h-0 min-w-0 flex-1 flex-col rounded-none border-0 shadow-none"
          contentClassName="flex min-h-0 flex-1 flex-col overflow-auto p-6"
        >
          {selectedRole ? (
            <RoleMemberTable
              role={selectedRole}
              members={members}
              onRemoveMember={handleRemoveMember}
              onAddMember={() => setAddMemberOpen(true)}
            />
          ) : (
            <div className="flex h-full flex-col items-center justify-center gap-2">
              <Shield className="h-10 w-10 text-muted-foreground/30" strokeWidth={1.5} />
              <p className="text-sm text-muted-foreground">请选择一个角色</p>
            </div>
          )}
        </DataSection>
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
          onSearchMembers={searchMembers}
        />
      )}

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
    </PageShell>
  )
}
