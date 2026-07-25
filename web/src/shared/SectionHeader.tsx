import { FadeIn } from '@/shared/FadeIn'
import { cn } from '@/shared/cn'

interface SectionHeaderProps {
  title: string
  subtitle?: string
  align?: 'left' | 'center'
  className?: string
  badge?: string
  titleHighlight?: string
}

export function SectionHeader({
  title,
  subtitle,
  align = 'left',
  className,
  badge,
  titleHighlight,
}: SectionHeaderProps) {
  const isCenter = align === 'center'

  return (
    <FadeIn
      margin="-100px"
      className={cn(
        'mb-10 sm:mb-12',
        isCenter ? 'text-center max-w-3xl mx-auto mb-14 sm:mb-16' : 'max-w-3xl',
        className,
      )}
    >
      {badge ? (
        <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-white/80 border border-brand-200/80 shadow-sm mb-5">
          <span className="w-1.5 h-1.5 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(139,92,246,0.1)]" />
          <span className="text-xs font-semibold tracking-wide text-brand-700">{badge}</span>
        </div>
      ) : null}

      {isCenter ? (
        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-black text-ink-950 leading-[1.15] tracking-[-0.035em]">
          {title}
          {titleHighlight ? (
            <>
              <br />
              <span className="gradient-brand-text">{titleHighlight}</span>
            </>
          ) : null}
        </h2>
      ) : (
        <div className="flex items-center gap-4">
          <div className="w-1 h-9 bg-gradient-to-b from-brand-500 to-cyan-500 rounded-full shrink-0 shadow-[0_0_20px_rgba(139,92,246,0.35)]" />
          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-black text-ink-950 tracking-[-0.03em]">
            {title}
          </h2>
        </div>
      )}

      {subtitle ? (
        <p
          className={cn(
            'mt-5 text-base sm:text-lg text-ink-600 leading-relaxed',
            isCenter ? '' : 'max-w-2xl',
          )}
        >
          {subtitle}
        </p>
      ) : null}
    </FadeIn>
  )
}
