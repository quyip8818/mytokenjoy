import type { ReactNode } from 'react'

export function MyStatGroup({
  title,
  icon: Icon,
  items,
  action,
}: {
  title: string
  icon: React.ComponentType<{ className?: string }>
  items: {
    label: string
    value: string
    icon: React.ComponentType<{ className?: string }>
    action?: ReactNode
  }[]
  action?: ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-5 shadow-xs">
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon className="size-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">{title}</h3>
        </div>
        {action}
      </div>
      <div className="space-y-3">
        {items.map((item) => (
          <div key={item.label} className="flex items-center gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted">
              <item.icon className="size-4 text-muted-foreground" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-xs text-muted-foreground">{item.label}</p>
              <p className="text-base font-semibold tabular-nums">{item.value}</p>
            </div>
            {item.action}
          </div>
        ))}
      </div>
    </div>
  )
}
