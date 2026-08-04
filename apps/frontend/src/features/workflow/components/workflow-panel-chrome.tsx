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
      <header className="flex shrink-0 items-center gap-3 border-b border-border px-5 py-4">
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
        <div className="shrink-0 border-b border-border/40 bg-muted/50 px-5 py-2 text-sm text-muted-foreground">
          {contextBar}
        </div>
      )}

      <div
        className={[
          'flex-1 overflow-y-auto px-10 py-8 text-base',
          // ponytail: 内容区限宽居中，宽面板不会太空；窄面板无影响
          '[&>*]:mx-auto [&>*]:w-full [&>*]:max-w-2xl',
          // ponytail: 统一放大面板内 input/select/textarea 尺寸，避免逐组件改 size prop
          '[&_input]:h-11 [&_input]:text-base',
          '[&_button[role=combobox]]:h-11 [&_button[role=combobox]]:text-base',
          '[&_textarea]:text-base',
        ].join(' ')}
      >
        {children}
      </div>

      {banner && (
        <div className="shrink-0 border-t border-amber-200/60 bg-amber-50/80 px-5 py-3 dark:border-amber-900/40 dark:bg-amber-950/30">
          {banner}
        </div>
      )}

      {footer && (
        <footer className="flex shrink-0 items-center justify-end gap-3 border-t border-border px-5 py-5 [&_button]:h-11 [&_button]:px-6 [&_button]:text-base">
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
  primaryDisabledReason?: string
  secondaryLabel?: string
  onSecondary?: () => void
  destructiveLabel?: string
  onDestructive?: () => void
  destructiveDisabled?: boolean
  destructiveDisabledReason?: string
}

export function WorkflowPanelFooter({
  onCancel,
  cancelLabel = '取消',
  primaryLabel,
  onPrimary,
  primaryDisabled,
  primaryDisabledReason,
  secondaryLabel,
  onSecondary,
  destructiveLabel,
  onDestructive,
  destructiveDisabled,
  destructiveDisabledReason,
}: WorkflowPanelFooterProps) {
  // ponytail: 统一兜底——disabled 但无 reason 时给一个通用提示，避免用户困惑。
  // 升级路径：各 workflow 逐步传入具体 reason 覆盖此默认值。
  const effectivePrimaryReason =
    primaryDisabledReason ?? (primaryDisabled ? '请完善必填信息' : undefined)
  const effectiveDestructiveReason =
    destructiveDisabledReason ?? (destructiveDisabled ? '当前无法执行此操作' : undefined)

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
        <Button
          variant="destructive"
          disabled={destructiveDisabled}
          disabledReason={effectiveDestructiveReason}
          onClick={onDestructive}
        >
          {destructiveLabel}
        </Button>
      )}
      <Button
        disabled={primaryDisabled}
        disabledReason={effectivePrimaryReason}
        variant="brand"
        onClick={onPrimary}
      >
        {primaryLabel}
      </Button>
    </>
  )
}
