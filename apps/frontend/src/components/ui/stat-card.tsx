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
  className?: string
}

export function StatCard({
  label,
  value,
  subValue,
  accent,
  icon: Icon,
  iconAccent = 'from-blue-500 to-sky-500',
  className,
}: StatCardProps) {
  return (
    <Card
      className={cn(
        'border-border/50 shadow-card transition-all duration-200',
        Icon && 'hover:shadow-card-hover hover:-translate-y-0.5',
        className,
      )}
    >
      <CardContent className={cn('pt-4 pb-3', Icon ? 'px-5' : undefined)}>
        {Icon ? (
          <div className="mb-3 flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{label}</span>
            <div
              className={cn(
                'flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br',
                iconAccent,
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
