import { cn } from '@/shared/cn'

interface LogoProps {
  size?: 'sm' | 'md' | 'lg'
  showText?: boolean
  className?: string
}

const SIZE_MAP = {
  sm: { wrap: 32, text: 'text-lg' },
  md: { wrap: 48, text: 'text-2xl' },
  lg: { wrap: 64, text: 'text-3xl' },
} as const

export function Logo({ size = 'md', showText = true, className }: LogoProps) {
  const { wrap, text } = SIZE_MAP[size]
  return (
    <div className={cn('flex items-center gap-2.5', className)}>
      <img
        src="/logo.png"
        alt="Tokenjoy Logo"
        width={wrap}
        height={wrap}
        className="shrink-0"
        style={{ width: wrap, height: wrap }}
      />
      {showText ? (
        <span
          className={cn('font-black text-ink-950 tracking-tight', text)}
          style={{ letterSpacing: '-0.03em' }}
        >
          Tokenjoy
        </span>
      ) : null}
    </div>
  )
}
