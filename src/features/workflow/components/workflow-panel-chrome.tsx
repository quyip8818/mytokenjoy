import type { ReactNode } from 'react'
import { ArrowLeft, X } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface WorkflowPanelChromeProps {
  title: string
  showBack?: boolean
  onBack?: () => void
  onClose: () => void
  contextBar?: ReactNode
  banner?: ReactNode
  footer?: ReactNode
  children: ReactNode
}

export function WorkflowPanelChrome({
  title,
  showBack,
  onBack,
  onClose,
  contextBar,
  banner,
  footer,
  children,
}: WorkflowPanelChromeProps) {
  return (
    <div className="flex h-full flex-col bg-card">
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-border/60 px-6">
        {showBack && onBack ? (
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onBack}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
        ) : (
          <div className="w-8" />
        )}
        <h2 className="flex-1 text-base font-semibold text-foreground">{title}</h2>
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </header>

      {contextBar && (
        <div className="shrink-0 border-b border-border/40 bg-muted/50 px-6 py-2 text-sm text-muted-foreground">
          {contextBar}
        </div>
      )}

      <div className="flex-1 overflow-y-auto px-6 py-5">{children}</div>

      {banner && (
        <div className="shrink-0 border-t border-amber-200/60 bg-amber-50/80 px-6 py-3 dark:border-amber-900/40 dark:bg-amber-950/30">
          {banner}
        </div>
      )}

      {footer && (
        <footer className="flex h-16 shrink-0 items-center justify-end gap-3 border-t border-border/60 px-6">
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
  secondaryLabel?: string
  onSecondary?: () => void
  destructiveLabel?: string
  onDestructive?: () => void
  destructiveDisabled?: boolean
}

export function WorkflowPanelFooter({
  onCancel,
  cancelLabel = '取消',
  primaryLabel,
  onPrimary,
  primaryDisabled,
  secondaryLabel,
  onSecondary,
  destructiveLabel,
  onDestructive,
  destructiveDisabled,
}: WorkflowPanelFooterProps) {
  return (
    <>
      {onCancel && (
        <Button variant="outline" onClick={onCancel}>
          {cancelLabel}
        </Button>
      )}
      {onSecondary && secondaryLabel && (
        <Button variant="outline" onClick={onSecondary}>
          {secondaryLabel}
        </Button>
      )}
      {onDestructive && destructiveLabel && (
        <Button variant="destructive" disabled={destructiveDisabled} onClick={onDestructive}>
          {destructiveLabel}
        </Button>
      )}
      <Button disabled={primaryDisabled} variant="brand" onClick={onPrimary}>
        {primaryLabel}
      </Button>
    </>
  )
}
