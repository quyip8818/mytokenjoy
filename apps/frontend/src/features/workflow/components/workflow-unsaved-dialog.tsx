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

interface WorkflowUnsavedDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}

export function WorkflowUnsavedDialog({ open, onCancel, onConfirm }: WorkflowUnsavedDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={(v) => !v && onCancel()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>放弃未保存的更改？</AlertDialogTitle>
          <AlertDialogDescription>当前面板有未保存的更改，关闭后将丢失。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onCancel}>继续编辑</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm}>放弃更改</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
