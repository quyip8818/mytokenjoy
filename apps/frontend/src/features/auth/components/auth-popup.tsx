import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { AuthCard, type AuthCardProps } from './auth-card'

interface AuthPopupProps extends AuthCardProps {
  open: boolean
  closable?: boolean
  onClose?: () => void
}

export function AuthPopup({
  open,
  defaultMode = 'login',
  closable = false,
  onSuccess,
  onClose,
}: AuthPopupProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={
        closable
          ? (v) => {
              if (!v) onClose?.()
            }
          : undefined
      }
    >
      <DialogContent
        className="sm:max-w-[480px] gap-0 p-0 overflow-hidden border-border/50 shadow-[0_10px_50px_rgba(139,92,246,0.12)]"
        onPointerDownOutside={closable ? undefined : (e) => e.preventDefault()}
        onEscapeKeyDown={closable ? undefined : (e) => e.preventDefault()}
        showCloseButton={closable}
      >
        <DialogTitle className="sr-only">TokenJoy 认证</DialogTitle>
        <AuthCard defaultMode={defaultMode} onSuccess={onSuccess} />
      </DialogContent>
    </Dialog>
  )
}
