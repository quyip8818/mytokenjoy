import type { ReactNode } from 'react'

export function MyChartSection({
  title,
  icon: Icon,
  children,
}: {
  title: string
  icon: React.ComponentType<{ className?: string }>
  children: ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card shadow-xs">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <Icon className="size-4 text-muted-foreground" />
        <h3 className="text-sm font-semibold">{title}</h3>
      </div>
      <div className="p-5">{children}</div>
    </div>
  )
}
