import { Shield } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import type { useRolesPage } from '@/features/org'
import { RoleList } from './role-list'
import { RoleForm } from './role-form'
import { RoleMemberTable, AddMemberDialog } from './role-member-table'

type RolesPageShellProps = ReturnType<typeof useRolesPage>

export function RolesPageShell({
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
}: RolesPageShellProps) {
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

      <ConfirmActionDialog
        state={
          deleteConfirm
            ? {
                open: true,
                title: '删除角色',
                desc:
                  deleteConfirm.memberCount > 0
                    ? `该角色下有 ${deleteConfirm.memberCount} 名成员，删除后将失去对应权限，是否继续？`
                    : '确定要删除该角色吗？',
                variant: 'danger',
                confirmLabel: '删除',
                onConfirm: handleConfirmDelete,
              }
            : null
        }
        onOpenChange={(open) => {
          if (!open) setDeleteConfirm(null)
        }}
        onClose={() => setDeleteConfirm(null)}
      />

      <ConfirmActionDialog
        state={
          removeConfirm
            ? {
                open: true,
                title: '移除成员',
                desc: `确定将「${removeConfirm.member.name}」从「${removeConfirm.role.name}」角色中移除吗？`,
                variant: 'danger',
                confirmLabel: '移除',
                onConfirm: handleConfirmRemove,
              }
            : null
        }
        onOpenChange={(open) => {
          if (!open) setRemoveConfirm(null)
        }}
        onClose={() => setRemoveConfirm(null)}
      />
    </PageShell>
  )
}
