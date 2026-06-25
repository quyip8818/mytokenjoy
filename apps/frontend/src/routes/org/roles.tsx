import { useState } from 'react'
import type { Member, Role } from '@/api/types'
import { roleApi } from '@/api/org'
import { RoleList } from '@/components/org/role-list'
import { RoleMemberTable } from '@/components/org/role-member-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { EmptyState } from '@/components/ui/empty-state'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoRole } from '@/features/demo'
import { usePermissions } from '@/hooks/use-permissions'
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
  const { memberId, refreshSession } = useDemoRole()
  const { canWrite } = usePermissions()
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Role | null>(null)
  const [removeConfirm, setRemoveConfirm] = useState<{ member: Member; role: Role } | null>(null)

  const {
    data: initData,
    loading,
    setData: setInitData,
  } = useAsyncResource(async () => {
    const [rolesData, permsData] = await Promise.all([roleApi.list(), roleApi.getPermissions()])
    return { roles: rolesData, permissions: permsData }
  }, [])

  const roles = initData?.roles ?? []
  const permissions = initData?.permissions ?? []
  const activeRoleId = selectedRoleId ?? roles[0]?.id ?? null

  const {
    data: members = [],
    loading: membersLoading,
    refresh: refreshMembers,
  } = useAsyncResource(async () => {
    if (!activeRoleId) return []
    return roleApi.getMembers(activeRoleId)
  }, [activeRoleId])

  const selectedRole = roles.find((r) => r.id === activeRoleId) ?? null

  const refreshRoles = async () => {
    const updated = await roleApi.list()
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
    await roleApi.removeMember(removeConfirm.role.id, removeConfirm.member.id)
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
