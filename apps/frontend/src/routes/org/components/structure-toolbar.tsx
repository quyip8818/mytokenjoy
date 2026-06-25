import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { PermissionGate } from '@/components/auth/permission-gate'

interface StructureToolbarProps {
  directOnly: boolean
  selectedCount: number
  onDirectOnlyChange: (directOnly: boolean) => void
  onBatchTransfer: () => void
  onBatchActivate: () => void
  onBatchDeactivate: () => void
  onBatchDelete: () => void
  onImportMembers: () => void
  onInviteMember: () => void
  onAddMember: () => void
}

export function StructureToolbar({
  directOnly,
  selectedCount,
  onDirectOnlyChange,
  onBatchTransfer,
  onBatchActivate,
  onBatchDeactivate,
  onBatchDelete,
  onImportMembers,
  onInviteMember,
  onAddMember,
}: StructureToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <Select
        value={directOnly ? 'direct' : 'all'}
        onValueChange={(v) => onDirectOnlyChange(v === 'direct')}
      >
        <SelectTrigger className="w-[100px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">全部</SelectItem>
          <SelectItem value="direct">仅直属</SelectItem>
        </SelectContent>
      </Select>

      {selectedCount > 0 && (
        <PermissionGate write>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" size="sm" />}>
              批量操作 ({selectedCount})
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={onBatchTransfer}>批量转移部门</DropdownMenuItem>
              <DropdownMenuItem onClick={onBatchActivate}>批量启用</DropdownMenuItem>
              <DropdownMenuItem onClick={onBatchDeactivate}>批量停用</DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onClick={onBatchDelete}>
                批量删除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </PermissionGate>
      )}

      <div className="flex-1" />
      <PermissionGate write>
        <Button variant="outline" size="sm" onClick={onImportMembers}>
          导入成员
        </Button>
        <Button variant="outline" size="sm" onClick={onInviteMember}>
          邀请成员
        </Button>
        <Button size="sm" variant="brand" onClick={onAddMember}>
          添加成员
        </Button>
      </PermissionGate>
    </div>
  )
}
