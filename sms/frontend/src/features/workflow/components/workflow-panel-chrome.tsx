import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface WorkflowPanelChromeProps {
  title: string
  onClose: () => void
  footer?: ReactNode
  children: ReactNode
}

export function WorkflowPanelChrome({
  title,
  onClose,
  footer,
  children,
}: WorkflowPanelChromeProps) {
  return (
    <div className="flex h-full flex-col bg-card">
      <header className="flex shrink-0 items-center gap-3 border-b border-border px-5 py-4">
        <h2 className="flex-1 text-base font-semibold text-foreground">{title}</h2>
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </header>

      <div className="flex-1 overflow-y-auto px-10 py-8 [&>*]:mx-auto [&>*]:w-full [&>*]:max-w-2xl">
        {children}
      </div>

      {footer && (
        <footer className="flex shrink-0 items-center justify-end gap-3 border-t border-border px-5 py-5">
          {footer}
        </footer>
      )}
    </div>
  )
}

interface WorkflowPanelFooterProps {
  onCancel?: () => void
  cancelLabel?: string
  primaryLabel: string
  onPrimary: () => void
  primaryDisabled?: boolean
}

export function WorkflowPanelFooter({
  onCancel,
  cancelLabel = '取消',
  primaryLabel,
  onPrimary,
  primaryDisabled,
}: WorkflowPanelFooterProps) {
  return (
    <>
      {onCancel && (
        <Button variant="outline" onClick={onCancel}>
          {cancelLabel}
        </Button>
      )}
      <Button disabled={primaryDisabled} onClick={onPrimary}>
        {primaryLabel}
      </Button>
    </>
  )
}
