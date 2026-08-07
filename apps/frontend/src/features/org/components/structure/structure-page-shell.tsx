import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { ContextHeader } from '@/components/layout/context-header'
import { DataSection } from '@/components/layout/data-section'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import type { useStructurePage } from '@/features/org'
import { DepartmentPanel } from './department-panel'
import { MemberFormDialog } from './member-form-dialog'
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
  setPageSize,
  setRowSelection,
  refreshDepartments,
  refreshMembers,
  setTransferOpen,
  setTransferDeptId,
  setConfirmState,
  handleMemberSubmit,
  handleSearch,
  handleStatusChange,
  handleDelete,
  handleBatchTransfer,
  openCreateMember,
  openEditMember,
  closeMemberForm,
}: StructurePageShellProps) {
  return (
    <PageShell testId="page-org-structure" className="flex min-h-0 flex-1 flex-col">
      <SplitPanel
        master={
          <DataSection
            loading={departmentsLoading}
            error={departmentsError}
            onRetry={() => void refreshDepartments()}
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
        }
        detail={
          <>
            {selectedDept && <ContextHeader title={selectedDept.name} />}
            <div className="min-h-0 flex-1 overflow-hidden">
              <DataSection
                loading={membersLoading}
                error={membersError}
                onRetry={() => void refreshMembers()}
                loadingVariant="spinner"
                className="flex h-full flex-col gap-4 overflow-hidden p-5"
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
                  onSearch={handleSearch}
                  onAdd={openCreateMember}
                  onPageChange={setPage}
                  onPageSizeChange={setPageSize}
                  onEdit={openEditMember}
                  onStatusChange={handleStatusChange}
                  onDelete={handleDelete}
                  onRowSelectionChange={setRowSelection}
                  onTransfer={() => setTransferOpen(true)}
                  onClearSelection={() => setRowSelection({})}
                />
              </DataSection>
            </div>
          </>
        }
      />

      <MemberFormDialog
        open={formOpen}
        member={editingMember}
        departments={departments}
        onSubmit={handleMemberSubmit}
        onClose={closeMemberForm}
      />

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
