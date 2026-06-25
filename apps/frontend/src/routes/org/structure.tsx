import { Users } from 'lucide-react'
import type { Member } from '@/api/types'
import { DepartmentTree } from '@/routes/org/components/department-tree'
import { MemberTable } from '@/routes/org/components/member-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { useStructurePage } from '@/routes/org/hooks/use-structure-page'
import { StructureSummaryCard } from '@/routes/org/components/structure-summary-card'
import { StructureToolbar } from '@/routes/org/components/structure-toolbar'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'

export default function StructurePage() {
  const {
    canWrite,
    selectedDept,
    departments,
    members,
    total,
    page,
    pageSize,
    membersLoading,
    error,
    refresh,
    directOnly,
    rowSelection,
    confirmState,
    inactiveCount,
    selectedIds,
    approvalPendingCount,
    refreshMembers,
    handleSelectDept,
    openMemberForm,
    handleStatusChange,
    handleDelete,
    handleBatchInvite,
    handleBatchTransfer,
    handleInvite,
    setDirectOnlyFilter,
    setPage,
    setRowSelection,
    closeConfirm,
    open,
    handleAddChildDept,
    handleUpdateDeptName,
    handleDeleteDept,
  } = useStructurePage()

  return (
    <PageShell
      layout="split"
      sidebar={
        <DepartmentTree
          departments={departments}
          selectedId={selectedDept?.id}
          onSelect={handleSelectDept}
          onAddChildDept={handleAddChildDept}
          onUpdateDeptName={handleUpdateDeptName}
          onDeleteDept={handleDeleteDept}
          readOnly={!canWrite}
        />
      }
    >
      <StructureSummaryCard
        selectedDept={selectedDept}
        total={total}
        inactiveCount={inactiveCount}
        approvalPendingCount={approvalPendingCount}
        onBatchInvite={handleBatchInvite}
      />

      <StructureToolbar
        directOnly={directOnly}
        selectedCount={selectedIds.length}
        onDirectOnlyChange={setDirectOnlyFilter}
        onBatchTransfer={handleBatchTransfer}
        onBatchActivate={() => handleStatusChange(selectedIds, 'active')}
        onBatchDeactivate={() => handleStatusChange(selectedIds, 'inactive')}
        onBatchDelete={() => handleDelete(selectedIds)}
        onImportMembers={() =>
          open('member-import', {
            defaultDeptName: selectedDept?.name,
            onSuccess: refreshMembers,
          })
        }
        onInviteMember={() =>
          open('member-invite', {
            onSubmit: handleInvite,
          })
        }
        onAddMember={() => openMemberForm(null)}
      />

      <DataSection
        loading={membersLoading}
        error={error}
        onRetry={refresh}
        skeletonColumns={6}
        empty={
          !membersLoading && members.length === 0 && total === 0
            ? {
                icon: Users,
                title: '暂无成员',
                description: selectedDept
                  ? `${selectedDept.name} 下还没有成员`
                  : '请先选择部门或添加成员',
                actionLabel: canWrite ? '添加成员' : undefined,
                onAction: canWrite ? () => openMemberForm(null) : undefined,
              }
            : null
        }
      >
        <MemberTable
          data={members}
          total={total}
          page={page}
          pageSize={pageSize}
          onPageChange={setPage}
          onEdit={(m: Member) => openMemberForm(m)}
          onStatusChange={handleStatusChange}
          onDelete={handleDelete}
          rowSelection={rowSelection}
          onRowSelectionChange={setRowSelection}
          readOnly={!canWrite}
        />
      </DataSection>

      <ConfirmActionDialog
        state={confirmState}
        onOpenChange={(isOpen) => {
          if (!isOpen) closeConfirm()
        }}
        onClose={closeConfirm}
      />
    </PageShell>
  )
}
