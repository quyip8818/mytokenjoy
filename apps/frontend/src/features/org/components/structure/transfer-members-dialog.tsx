import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface TransferMembersDialogProps {
  open: boolean
  transferDeptId: string
  flatDepts: { id: string; name: string; level: number }[]
  onOpenChange: (open: boolean) => void
  onDeptChange: (deptId: string) => void
  onConfirm: () => void
  onCancel: () => void
}

export function TransferMembersDialog({
  open,
  transferDeptId,
  flatDepts,
  onOpenChange,
  onDeptChange,
  onConfirm,
  onCancel,
}: TransferMembersDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>批量转移部门</DialogTitle>
        </DialogHeader>
        <Select value={transferDeptId} onValueChange={(v) => onDeptChange(v ?? '')}>
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
          <Button variant="outline" onClick={onCancel}>
            取消
          </Button>
          <Button onClick={onConfirm}>确定转移</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
