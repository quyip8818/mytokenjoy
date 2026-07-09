import type { RowSelectionState } from '@tanstack/react-table'
import type { Member } from '@/api/types'
import { MemberToolbar } from './member-toolbar'
import { MemberTable } from './member-table'
import { BatchActionBar } from './batch-action-bar'
import { PendingActivationBanner } from './pending-activation-banner'

interface StructureMembersPanelProps {
  selectedDeptName?: string
  members: Member[]
  total: number
  page: number
  pageSize: number
  keyword: string
  rowSelection: RowSelectionState
  pendingCount: number
  selectedIds: string[]
  onKeywordChange: (keyword: string) => void
  onSearch: () => void
  onInvite: () => void
  onAdd: () => void
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  onEdit: (member: Member) => void
  onStatusChange: (ids: string[], status: 'active' | 'inactive') => void
  onDelete: (ids: string[]) => void
  onRowSelectionChange: (selection: RowSelectionState) => void
  onTransfer: () => void
  onClearSelection: () => void
}

export function StructureMembersPanel({
  selectedDeptName,
  members,
  total,
  page,
  pageSize,
  keyword,
  rowSelection,
  pendingCount,
  selectedIds,
  onKeywordChange,
  onSearch,
  onInvite,
  onAdd,
  onPageChange,
  onPageSizeChange,
  onEdit,
  onStatusChange,
  onDelete,
  onRowSelectionChange,
  onTransfer,
  onClearSelection,
}: StructureMembersPanelProps) {
  return (
    <>
      <div className="flex items-center gap-4">
        <h3 className="text-sm font-semibold text-foreground">{selectedDeptName ?? '全部成员'}</h3>
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

      <PendingActivationBanner pendingCount={pendingCount} />

      <MemberToolbar
        keyword={keyword}
        onKeywordChange={onKeywordChange}
        onSearch={onSearch}
        onInvite={onInvite}
        onAdd={onAdd}
      />

      <MemberTable
        data={members}
        total={total}
        page={page}
        pageSize={pageSize}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
        onEdit={onEdit}
        onStatusChange={onStatusChange}
        onDelete={onDelete}
        rowSelection={rowSelection}
        onRowSelectionChange={onRowSelectionChange}
      />

      <BatchActionBar
        count={selectedIds.length}
        onTransfer={onTransfer}
        onEnable={() => onStatusChange(selectedIds, 'active')}
        onDisable={() => onStatusChange(selectedIds, 'inactive')}
        onDelete={() => onDelete(selectedIds)}
        onClear={onClearSelection}
      />
    </>
  )
}
