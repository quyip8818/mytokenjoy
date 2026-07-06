import { useEffect, useState } from 'react'
import { useOrgStructureStore } from '@/stores/org-structure'
import type { Department, Member } from '@/api/types'
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
import { DepartmentPanel } from '@/components/org/structure/department-panel'
import { MemberToolbar } from '@/components/org/structure/member-toolbar'
import { MemberTable } from '@/components/org/structure/member-table'
import { MemberFormDialog } from '@/components/org/structure/member-form-dialog'
import { BatchActionBar } from '@/components/org/structure/batch-action-bar'
import { InviteDialog } from '@/components/org/structure/invite-dialog'
import { AlertTriangle, Send } from 'lucide-react'

export default function StructurePage() {
  const store = useOrgStructureStore()

  const [formOpen, setFormOpen] = useState(false)
  const [editingMember, setEditingMember] = useState<Member | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [transferDeptId, setTransferDeptId] = useState('')
  const [confirmState, setConfirmState] = useState<{
    open: boolean; title: string; desc: string; variant: 'primary' | 'danger'; onConfirm: () => void
  }>({ open: false, title: '', desc: '', variant: 'primary', onConfirm: () => {} })

  useEffect(() => {
    store.loadDepartments()
    store.loadMembers()
  }, [])

  const pendingCount = store.members.filter((m) => m.status === 'pending').length
  const selectedIds = Object.keys(store.rowSelection)

  const handleMemberSubmit = async (data: {
    name: string; phone: string; email: string; username: string
    employeeId: string; jobTitle: string; hireDate: string; departmentId: string
  }) => {
    if (editingMember) {
      await store.updateMember(editingMember.id, data)
    } else {
      const dept = flattenDepts(store.departments).find((d) => d.id === data.departmentId)
      await store.createMember({ ...data, departmentName: dept?.name ?? '' })
    }
    setFormOpen(false)
    setEditingMember(null)
  }

  const handleStatusChange = (ids: string[], status: 'active' | 'inactive') => {
    setConfirmState({
      open: true,
      title: status === 'inactive' ? '停用成员' : '启用成员',
      desc: status === 'inactive'
        ? '停用后该成员的 Platform Key 将同步失效'
        : `确定启用选中的 ${ids.length} 名成员？`,
      variant: status === 'inactive' ? 'danger' : 'primary',
      onConfirm: async () => {
        await store.updateMemberStatus(ids, status)
        setConfirmState((s) => ({ ...s, open: false }))
      },
    })
  }

  const handleDelete = (ids: string[]) => {
    setConfirmState({
      open: true,
      title: '删除成员',
      desc: `确定删除 ${ids.length} 名成员？此操作不可恢复`,
      variant: 'danger',
      onConfirm: async () => {
        await store.deleteMember(ids)
        setConfirmState((s) => ({ ...s, open: false }))
      },
    })
  }

  const handleBatchTransfer = async () => {
    if (!transferDeptId || selectedIds.length === 0) return
    await store.transferMembers(selectedIds, transferDeptId)
    setTransferOpen(false)
    setTransferDeptId('')
  }

  const flatDepts = flattenDepts(store.departments)

  return (
    <div className="flex h-[calc(100dvh-7.5rem)] rounded-lg border border-border bg-card shadow-xs overflow-hidden">
      {/* Left: Department Tree */}
      <DepartmentPanel
        selectedId={store.selectedDept?.id}
        onSelect={store.selectDept}
        onTreeChange={store.loadDepartments}
      />

      {/* Right: Member Content */}
      <div className="relative flex flex-1 flex-col gap-4 overflow-hidden p-5">
        {/* Stats header */}
        <div className="flex items-center gap-4">
          <h3 className="text-sm font-semibold text-foreground">
            {store.selectedDept?.name ?? '全部成员'}
          </h3>
          <div className="h-4 w-px bg-border" />
          <span className="text-xs text-muted-foreground">
            共 <span className="font-medium tabular-nums text-foreground">{store.total}</span> 人
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

        {/* Pending activation banner */}
        {pendingCount > 0 && (
          <div className="flex items-center gap-3 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-800">
            <AlertTriangle className="size-4 shrink-0 text-amber-600" />
            <span className="flex-1">
              当前有 <span className="font-medium">{pendingCount}</span> 名成员尚未激活
            </span>
            <Button variant="ghost" size="sm" className="h-7 text-xs text-amber-700 hover:bg-amber-100">
              <Send className="size-3.5" />
              发送激活邀请
            </Button>
          </div>
        )}

        {/* Toolbar */}
        <MemberToolbar
          keyword={store.keyword}
          onKeywordChange={store.setKeyword}
          onInvite={() => setInviteOpen(true)}
          onAdd={() => { setEditingMember(null); setFormOpen(true) }}
        />

        {/* Table */}
        <MemberTable
          data={store.members}
          total={store.total}
          page={store.page}
          pageSize={store.pageSize}
          onPageChange={store.setPage}
          onEdit={(m) => { setEditingMember(m); setFormOpen(true) }}
          onStatusChange={handleStatusChange}
          onDelete={handleDelete}
          rowSelection={store.rowSelection}
          onRowSelectionChange={store.setRowSelection}
        />

        {/* Batch action bar */}
        <BatchActionBar
          count={selectedIds.length}
          onTransfer={() => setTransferOpen(true)}
          onEnable={() => handleStatusChange(selectedIds, 'active')}
          onDisable={() => handleStatusChange(selectedIds, 'inactive')}
          onDelete={() => handleDelete(selectedIds)}
          onClear={() => store.setRowSelection({})}
        />
      </div>

      {/* Dialogs */}
      <MemberFormDialog
        open={formOpen}
        member={editingMember}
        departments={store.departments}
        onSubmit={handleMemberSubmit}
        onClose={() => { setFormOpen(false); setEditingMember(null) }}
      />

      <InviteDialog open={inviteOpen} onOpenChange={setInviteOpen} onInvite={store.inviteMember} />

      <Dialog open={transferOpen} onOpenChange={(open) => { if (!open) setTransferOpen(false) }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader><DialogTitle>批量转移部门</DialogTitle></DialogHeader>
          <Select value={transferDeptId} onValueChange={(v) => setTransferDeptId(v ?? '')}>
            <SelectTrigger className="w-full"><SelectValue placeholder="请选择目标部门" /></SelectTrigger>
            <SelectContent>
              {flatDepts.map((d) => (
                <SelectItem key={d.id} value={d.id}>{'　'.repeat(d.level)}{d.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTransferOpen(false)}>取消</Button>
            <Button onClick={handleBatchTransfer}>确定转移</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={confirmState.open} onOpenChange={(open) => { if (!open) setConfirmState((s) => ({ ...s, open: false })) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmState.title}</AlertDialogTitle>
            <AlertDialogDescription>{confirmState.desc}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setConfirmState((s) => ({ ...s, open: false }))}>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmState.onConfirm}
              className={confirmState.variant === 'danger' ? 'bg-destructive text-white hover:bg-destructive/90' : ''}
            >
              确认
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function flattenDepts(departments: Department[], level = 0): { id: string; name: string; level: number }[] {
  const result: { id: string; name: string; level: number }[] = []
  for (const dept of departments) {
    result.push({ id: dept.id, name: dept.name, level })
    if (dept.children) result.push(...flattenDepts(dept.children, level + 1))
  }
  return result
}
