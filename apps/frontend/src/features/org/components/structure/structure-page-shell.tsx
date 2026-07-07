import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import type { useStructurePage } from '@/features/org/hooks/use-structure-page'
import { DepartmentPanel } from './department-panel'
import { MemberFormDialog } from './member-form-dialog'
import { InviteDialog } from './invite-dialog'
import { TransferMembersDialog } from './transfer-members-dialog'
import { StructureMembersPanel } from './structure-members-panel'

type StructurePageShellProps = ReturnType<typeof useStructurePage>

export function StructurePageShell({
  departments,
  selectedDept,
  members,
  total,
  page,
  pageSize,
  keyword,
  rowSelection,
  departmentsLoading,
  departmentsError,
  membersLoading,
  membersError,
  formOpen,
  editingMember,
  inviteOpen,
  transferOpen,
  transferDeptId,
  confirmState,
  pendingCount,
  selectedIds,
  flatDepts,
  selectDept,
  createDept,
  updateDept,
  deleteDept,
  setKeyword,
  setPage,
  setRowSelection,
  refreshDepartments,
  refreshMembers,
  setInviteOpen,
  setTransferOpen,
  setTransferDeptId,
  setConfirmState,
  handleMemberSubmit,
  handleStatusChange,
  handleDelete,
  handleBatchTransfer,
  inviteMember,
  openCreateMember,
  openEditMember,
  closeMemberForm,
}: StructurePageShellProps) {
  return (
    <PageShell layout="fill">
      <div className="flex min-h-0 flex-1 overflow-hidden rounded-lg border border-border bg-card shadow-xs">
        <DataSection
          loading={departmentsLoading}
          error={departmentsError}
          onRetry={() => void refreshDepartments()}
          className="shrink-0 rounded-none border-0 shadow-none"
          contentClassName="h-full p-0"
          loadingVariant="spinner"
        >
          <DepartmentPanel
            tree={departments}
            selectedId={selectedDept?.id}
            onSelect={selectDept}
            onCreateDept={createDept}
            onUpdateDept={updateDept}
            onDeleteDept={deleteDept}
          />
        </DataSection>

        <DataSection
          loading={membersLoading}
          error={membersError}
          onRetry={() => void refreshMembers()}
          className="flex min-h-0 min-w-0 flex-1 flex-col rounded-none border-0 shadow-none"
          contentClassName="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden p-5"
        >
          <StructureMembersPanel
            selectedDeptName={selectedDept?.name}
            members={members}
            total={total}
            page={page}
            pageSize={pageSize}
            keyword={keyword}
            rowSelection={rowSelection}
            pendingCount={pendingCount}
            selectedIds={selectedIds}
            onKeywordChange={setKeyword}
            onInvite={() => setInviteOpen(true)}
            onAdd={openCreateMember}
            onPageChange={setPage}
            onEdit={openEditMember}
            onStatusChange={handleStatusChange}
            onDelete={handleDelete}
            onRowSelectionChange={setRowSelection}
            onTransfer={() => setTransferOpen(true)}
            onClearSelection={() => setRowSelection({})}
          />
        </DataSection>
      </div>

      <MemberFormDialog
        open={formOpen}
        member={editingMember}
        departments={departments}
        onSubmit={handleMemberSubmit}
        onClose={closeMemberForm}
      />

      <InviteDialog open={inviteOpen} onOpenChange={setInviteOpen} onInvite={inviteMember} />

      <TransferMembersDialog
        open={transferOpen}
        transferDeptId={transferDeptId}
        flatDepts={flatDepts}
        onOpenChange={(open) => {
          if (!open) setTransferOpen(false)
        }}
        onDeptChange={setTransferDeptId}
        onConfirm={handleBatchTransfer}
        onCancel={() => setTransferOpen(false)}
      />

      <ConfirmActionDialog
        state={confirmState.open ? confirmState : null}
        onOpenChange={(open) => {
          if (!open) setConfirmState((s) => ({ ...s, open: false }))
        }}
        onClose={() => setConfirmState((s) => ({ ...s, open: false }))}
      />
    </PageShell>
  )
}
