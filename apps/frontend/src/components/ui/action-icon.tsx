import type { ReactNode, ButtonHTMLAttributes } from 'react'
import { Tooltip, TooltipContent, TooltipTrigger } from './tooltip'
import { cn } from '@/lib/utils'

/** ponytail: shared class for action icon styling — reuse on non-button elements (e.g. Link) */
export const actionIconClass =
  'cursor-pointer rounded p-1.5 text-muted-foreground transition-transform hover:scale-110 active:scale-95 hover:bg-muted hover:text-foreground'

interface ActionIconProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** Tooltip hint text */
  hint: string
  /** Icon element (lucide-react etc.) */
  children: ReactNode
}

/**
 * ponytail: 通用表格操作 icon 按钮
 * - hover 放大 (scale-110) + cursor-pointer
 * - 自带 Tooltip hint
 * 升级路径：需要 loading/disabled 态时扩展 props
 */
export function ActionIcon({ hint, children, className, ...props }: ActionIconProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button type="button" className={cn(actionIconClass, className)} {...props}>
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent>{hint}</TooltipContent>
    </Tooltip>
  )
}
