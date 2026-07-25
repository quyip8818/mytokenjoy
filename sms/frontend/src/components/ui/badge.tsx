import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium',
  {
    variants: {
      variant: {
        default: 'bg-blue-50 text-blue-700 border-blue-200',
        success: 'bg-green-50 text-green-700 border-green-200',
        warning: 'bg-yellow-50 text-yellow-700 border-yellow-200',
        destructive: 'bg-red-50 text-red-700 border-red-200',
        outline: 'bg-gray-50 text-gray-600 border-gray-200',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />
}

/**
 * StatusBadge: renders a badge from a status enum map.
 * Usage: <StatusBadge status={row.status} map={SUPPLIER_STATUS} />
 */
export function StatusBadge({
  status,
  map,
}: {
  status: string
  map: Record<string, { label: string; variant: string }>
}) {
  const item = map[status]
  if (!item) return <span>{status}</span>
  return <Badge variant={item.variant as BadgeProps['variant']}>{item.label}</Badge>
}
