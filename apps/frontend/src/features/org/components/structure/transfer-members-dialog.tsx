import { FormDialog } from '@/components/ui/form-dialog'
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
    <FormDialog
      open={open}
      onOpenChange={(v) => {
        if (!v) onCancel()
        onOpenChange(v)
      }}
      title="批量转移部门"
      submitLabel="确定转移"
      submitDisabled={!transferDeptId}
      onSubmit={onConfirm}
      className="sm:max-w-sm"
    >
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
    </FormDialog>
  )
}
