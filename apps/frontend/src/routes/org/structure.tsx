import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
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
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import {
  DepartmentPanel,
  MemberToolbar,
  MemberTable,
  MemberFormDialog,
  BatchActionBar,
  InviteDialog,
  useStructurePage,
} from '@/features/org'
import { AlertTriangle, Send } from 'lucide-react'

export default function StructurePage() {
  const {
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
  } = useStructurePage()

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
          <div className="flex items-center gap-4">
            <h3 className="text-sm font-semibold text-foreground">
              {selectedDept?.name ?? '全部成员'}
            </h3>
            <div className="h-4 w-px bg-border" />
            <span className="text-xs text-muted-foreground">
              共 <span className="font-medium tabular-nums text-foreground">{total}</span> 人
            </span>
            {pendingCount > 0 && (
              <>
                <div className="h-4 w-px bg-border" />
                <span className="text-xs tabular-nums font-medium text-amber-600">
                  {pendingCount} 人待激活
                </span>
              </>
            )}
          </div>

          {pendingCount > 0 && (
            <div className="flex items-center gap-3 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-800">
              <AlertTriangle className="size-4 shrink-0 text-amber-600" />
              <span className="flex-1">
                当前有 <span className="font-medium">{pendingCount}</span> 名成员尚未激活
              </span>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs text-amber-700 hover:bg-amber-100"
              >
                <Send className="size-3.5" />
                发送激活邀请
              </Button>
            </div>
          )}

          <MemberToolbar
            keyword={keyword}
            onKeywordChange={setKeyword}
            onInvite={() => setInviteOpen(true)}
            onAdd={openCreateMember}
          />

          <MemberTable
            data={members}
            total={total}
            page={page}
            pageSize={pageSize}
            onPageChange={setPage}
            onEdit={openEditMember}
            onStatusChange={handleStatusChange}
            onDelete={handleDelete}
            rowSelection={rowSelection}
            onRowSelectionChange={setRowSelection}
          />

          <BatchActionBar
            count={selectedIds.length}
            onTransfer={() => setTransferOpen(true)}
            onEnable={() => handleStatusChange(selectedIds, 'active')}
            onDisable={() => handleStatusChange(selectedIds, 'inactive')}
            onDelete={() => handleDelete(selectedIds)}
            onClear={() => setRowSelection({})}
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

      <Dialog
        open={transferOpen}
        onOpenChange={(open) => {
          if (!open) setTransferOpen(false)
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>批量转移部门</DialogTitle>
          </DialogHeader>
          <Select value={transferDeptId} onValueChange={(v) => setTransferDeptId(v ?? '')}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="请选择目标部门" />
            </SelectTrigger>
            <SelectContent>
              {flatDepts.map((d) => (
                <SelectItem key={d.id} value={d.id}>
                  {'　'.repeat(d.level)}
                  {d.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTransferOpen(false)}>
              取消
            </Button>
            <Button onClick={handleBatchTransfer}>确定转移</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={confirmState.open}
        onOpenChange={(open) => {
          if (!open) setConfirmState((s) => ({ ...s, open: false }))
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmState.title}</AlertDialogTitle>
            <AlertDialogDescription>{confirmState.desc}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setConfirmState((s) => ({ ...s, open: false }))}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmState.onConfirm}
              className={
                confirmState.variant === 'danger'
                  ? 'bg-destructive text-white hover:bg-destructive/90'
                  : ''
              }
            >
              确认
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageShell>
  )
}
