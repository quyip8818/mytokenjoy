import { Shield } from 'lucide-react'
import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { ContextHeader } from '@/components/layout/context-header'
import { DataSection } from '@/components/layout/data-section'
import { EmptyState } from '@/components/ui/empty-state'
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
    <PageShell className="flex min-h-0 flex-1 flex-col">
      <SplitPanel
        master={
          <DataSection
            loading={rolesLoading}
            error={rolesError}
            onRetry={() => void refreshRoles()}
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
        }
        detail={
          <>
            {selectedRole && <ContextHeader title={selectedRole.name} />}
            <div className="min-h-0 flex-1 overflow-auto p-6">
              <DataSection
                loading={membersLoading}
                error={membersError}
                onRetry={() => void refreshMembers()}
                loadingVariant="spinner"
              >
                {selectedRole ? (
                  <RoleMemberTable
                    role={selectedRole}
                    members={members}
                    onRemoveMember={handleRemoveMember}
                    onAddMember={() => setAddMemberOpen(true)}
                  />
                ) : (
                  <EmptyState variant="minimal" icon={Shield} title="请选择一个角色" />
                )}
              </DataSection>
            </div>
          </>
        }
      />

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
                desc: `确定将「${removeConfirm.member.alias}」从「${removeConfirm.role.name}」角色中移除吗？`,
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
