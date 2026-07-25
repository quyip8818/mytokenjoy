import type { CSSProperties } from 'react'
import { ArrowRight, Play, Sparkles } from 'lucide-react'
import type { HeroContent } from '@/content/types'

export interface HeroProps {
  content: HeroContent
}

export function Hero({ content }: HeroProps) {
  return (
    <section className="relative min-h-[880px] lg:min-h-screen flex items-center justify-center overflow-hidden pt-24 pb-16 hero-ambient">
      <div className="absolute top-24 -left-24 w-80 h-80 bg-brand-400/20 rounded-full blur-3xl animate-float" />
      <div
        className="absolute top-32 -right-24 w-96 h-96 bg-cyan-400/15 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '2s' }}
      />
      <div
        className="absolute bottom-10 left-1/3 w-80 h-80 bg-orange-400/10 rounded-full blur-3xl animate-float"
        style={{ animationDelay: '4s' }}
      />
      <div className="absolute inset-0 bg-light-grid opacity-60" />
      <div className="absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-white via-white/70 to-transparent" />

      <div className="relative z-10 max-w-6xl mx-auto px-5 sm:px-6 text-center">
        <div
          className="inline-flex items-center gap-2.5 px-4 py-2 rounded-full glass-light mb-8 sm:mb-10 animate-fade-up"
          style={{ '--anim-delay': '0.1s', '--anim-duration': '0.6s' } as CSSProperties}
        >
          <span className="flex items-center justify-center w-5 h-5 rounded-full bg-brand-100">
            <Sparkles className="w-3 h-3 text-brand-600" />
          </span>
          <span className="text-xs sm:text-sm text-ink-700 font-semibold tracking-wide">
            {content.badge}
          </span>
          <ArrowRight className="w-3 h-3 text-ink-500" />
        </div>

        <h1
          className="mx-auto max-w-5xl text-[2.65rem] sm:text-6xl lg:text-[5.25rem] font-black text-ink-950 leading-[1.04] tracking-tight animate-fade-up"
          style={
            {
              letterSpacing: '-0.03em',
              '--anim-delay': '0.2s',
              '--anim-duration': '0.8s',
              '--anim-from-y': '30px',
            } as CSSProperties
          }
        >
          {content.title}
        </h1>

        <p
          className="mt-6 sm:mt-8 text-xl sm:text-3xl lg:text-4xl font-bold gradient-brand-text tracking-wide animate-fade-up"
          style={
            {
              letterSpacing: '0.05em',
              '--anim-delay': '0.4s',
              '--anim-duration': '0.6s',
            } as CSSProperties
          }
        >
          {content.subtitle}
        </p>

        <div
          className="mt-9 sm:mt-12 flex items-center justify-center animate-fade-up"
          style={{ '--anim-delay': '0.6s', '--anim-duration': '0.6s' } as CSSProperties}
        >
          <div className="flex items-center gap-4 sm:gap-5 max-w-2xl px-5 sm:px-7 py-4 rounded-2xl bg-white/50 border border-white/80 shadow-sm backdrop-blur-sm">
            <div className="w-1 h-10 bg-gradient-to-b from-brand-500 to-cyan-500 rounded-full shrink-0" />
            <p className="text-base sm:text-xl text-ink-700 font-semibold text-left">
              {content.quote}
            </p>
          </div>
        </div>

        <div
          className="mt-9 sm:mt-10 flex flex-col sm:flex-row items-center justify-center gap-3 animate-fade-up"
          style={{ '--anim-delay': '0.8s', '--anim-duration': '0.6s' } as CSSProperties}
        >
          <a
            href={content.primaryCta.href}
            className="group w-full sm:w-auto flex items-center justify-center gap-2 px-7 py-3.5 rounded-full bg-ink-950 text-white font-semibold text-sm sm:text-base hover:bg-brand-600 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-[0_14px_40px_rgba(91,33,182,0.25)]"
          >
            {content.primaryCta.label}
            <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
          </a>
          <a
            href={content.secondaryCta.href}
            className="group w-full sm:w-auto flex items-center justify-center gap-2 px-7 py-3.5 rounded-full bg-white/80 text-ink-950 font-semibold text-sm sm:text-base border border-ink-300/70 shadow-sm hover:border-brand-300 hover:bg-white transition-all duration-300"
          >
            <Play className="w-4 h-4 fill-current" />
            {content.secondaryCta.label}
          </a>
        </div>

        <div
          className="mt-14 sm:mt-16 pt-7 border-t border-ink-300/40 max-w-4xl mx-auto animate-fade-in"
          style={{ '--anim-delay': '1s', '--anim-duration': '0.6s' } as CSSProperties}
        >
          <p className="text-xs text-ink-500 uppercase tracking-widest font-medium mb-6">
            {content.trustLabel}
          </p>
          <div className="flex flex-wrap items-center justify-center gap-x-8 sm:gap-x-12 gap-y-4">
            {content.trustBrands.map((brand) => (
              <span
                key={brand}
                className="text-sm sm:text-base font-bold text-ink-600/75 tracking-wide"
              >
                {brand}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
