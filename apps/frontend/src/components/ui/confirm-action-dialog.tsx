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

export interface ConfirmActionState {
  open: boolean
  title: string
  desc: string
  variant: 'primary' | 'danger'
  confirmLabel?: string
  onConfirm: () => void
}

interface ConfirmActionDialogProps {
  state: ConfirmActionState | null
  onOpenChange: (open: boolean) => void
  onClose: () => void
}

export function ConfirmActionDialog({ state, onOpenChange, onClose }: ConfirmActionDialogProps) {
  if (!state) return null

  return (
    <AlertDialog open={state.open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{state.title}</AlertDialogTitle>
          <AlertDialogDescription>{state.desc}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={state.onConfirm}
            variant={state.variant === 'danger' ? 'destructive' : 'default'}
            className={
              state.variant === 'danger' && !state.confirmLabel
                ? 'bg-destructive text-destructive-foreground hover:bg-destructive/80'
                : undefined
            }
          >
            {state.confirmLabel ?? '确认'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
