import { Shield } from 'lucide-react'
import { RoleList } from '@/components/org/role-list'
import { RoleMemberTable } from '@/components/org/role-member-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { EmptyState } from '@/components/ui/empty-state'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { usePermissions } from '@/hooks/use-permissions'
import { useRolesPage } from '@/routes/org/hooks/use-roles-page'

export default function RolesPage() {
  const { canWrite } = usePermissions()
  const {
    roles,
    activeRoleId,
    selectedRole,
    members,
    loading,
    membersLoading,
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
  } = useRolesPage()

  return (
    <PageShell
      layout="split"
      sidebar={
        <DataSection
          loading={loading}
          skeletonColumns={1}
          skeletonRows={4}
          className="h-full w-[220px] shrink-0"
          contentClassName="p-4"
        >
          <RoleList
            roles={roles}
            selectedRoleId={activeRoleId}
            onSelect={handleSelectRole}
            onAdd={handleAddRole}
            onEdit={handleEditRole}
            onDelete={handleDeleteRole}
            readOnly={!canWrite}
          />
        </DataSection>
      }
    >
      <DataSection loading={membersLoading} skeletonColumns={4} className="min-h-0 flex-1">
        {selectedRole ? (
          <RoleMemberTable
            role={selectedRole}
            members={members}
            onRemoveMember={handleRemoveMember}
            onAddMember={handleAddMember}
            readOnly={!canWrite}
          />
        ) : (
          <EmptyState
            icon={Shield}
            title="请选择一个角色"
            description="从左侧列表选择角色查看成员，或创建新角色"
            actionLabel={canWrite ? '创建角色' : undefined}
            onAction={canWrite ? handleAddRole : undefined}
          />
        )}
      </DataSection>

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
