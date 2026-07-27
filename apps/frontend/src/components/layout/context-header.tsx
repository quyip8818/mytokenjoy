import { Fragment, type ReactNode } from 'react'
import { ChevronRight } from 'lucide-react'

export interface ContextHeaderProps {
  breadcrumb?: string[]
  title?: string
  actions?: ReactNode
}

export function ContextHeader({ breadcrumb, title, actions }: ContextHeaderProps) {
  const hasContent = breadcrumb?.length || title

  if (!hasContent && !actions) return null

  return (
    <div className="flex items-center justify-between border-b border-border px-5 py-3">
      <div className="flex items-center gap-1.5 text-xs">
        {breadcrumb?.map((segment, i) => (
          <Fragment key={i}>
            {i > 0 && <ChevronRight className="size-3 text-muted-foreground" aria-hidden />}
            <span
              className={
                i === breadcrumb.length - 1
                  ? 'font-medium text-foreground'
                  : 'text-muted-foreground'
              }
            >
              {segment}
            </span>
          </Fragment>
        ))}
        {title && !breadcrumb?.length && (
          <span className="font-medium text-foreground">{title}</span>
        )}
      </div>
      {actions && <div className="flex items-center gap-3">{actions}</div>}
    </div>
  )
}
