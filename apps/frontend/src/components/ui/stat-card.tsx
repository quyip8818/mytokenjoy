import type { LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface StatCardProps {
  label: string
  value: string
  subValue?: string
  accent?: boolean
  icon?: LucideIcon
  iconAccent?: string
  iconAccentStyle?: 'gradient' | 'solid'
  iconLayout?: 'corner' | 'inline'
  className?: string
}

export function StatCard({
  label,
  value,
  subValue,
  accent,
  icon: Icon,
  iconAccent = 'from-blue-500 to-sky-500',
  iconAccentStyle = 'gradient',
  iconLayout = 'corner',
  className,
}: StatCardProps) {
  const isSolid = iconAccentStyle === 'solid'
  const isInline = iconLayout === 'inline'

  if (isInline && Icon) {
    return (
      <Card className={cn('border-border bg-card shadow-xs', className)}>
        <CardContent className="p-4">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Icon className="size-3.5" strokeWidth={1.5} />
            {label}
          </div>
          <p
            className={cn(
              'mt-2 text-xl font-semibold tracking-tight tabular-nums',
              accent && 'text-primary',
            )}
          >
            {value}
          </p>
          {subValue ? <p className="mt-1 text-xs text-muted-foreground">{subValue}</p> : null}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card
      className={cn(
        isSolid ? 'border-border shadow-xs' : 'border-border/50 shadow-card',
        !isSolid &&
          Icon &&
          'transition-all duration-200 hover:-translate-y-0.5 hover:shadow-card-hover',
        className,
      )}
    >
      <CardContent
        className={cn(isSolid ? 'px-5 pt-5 pb-4' : cn('pt-4 pb-3', Icon ? 'px-5' : undefined))}
      >
        {Icon ? (
          <div className="mb-3 flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{label}</span>
            <div
              className={cn(
                'flex h-8 w-8 items-center justify-center rounded-lg',
                isSolid ? iconAccent : cn('bg-gradient-to-br', iconAccent),
              )}
            >
              <Icon className="h-4 w-4 text-white" />
            </div>
          </div>
        ) : (
          <p className="text-xs font-medium text-muted-foreground">{label}</p>
        )}
        <p
          className={cn(
            'font-bold tracking-tight tabular-nums',
            Icon ? 'text-2xl' : 'mt-1 text-2xl',
            accent && 'text-primary',
          )}
        >
          {value}
        </p>
        {subValue ? <p className="mt-1 text-xs text-muted-foreground">{subValue}</p> : null}
      </CardContent>
    </Card>
  )
}
