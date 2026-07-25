import { ArrowRight, Check, Sparkles } from 'lucide-react'
import { FadeIn } from '@/shared/FadeIn'
import type { CtaContent } from '@/content/types'

export interface DownloadCTAProps {
  content: CtaContent
}

export function DownloadCTA({ content }: DownloadCTAProps) {
  return (
    <section id="cta" className="relative py-20 sm:py-24 lg:py-32 overflow-hidden">
      <div className="absolute inset-0 bg-light-grid opacity-30" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[500px] bg-brand-300/20 rounded-full blur-[120px] animate-glow-pulse" />
      <div className="absolute top-1/3 right-1/4 w-[400px] h-[400px] bg-cyan-300/20 rounded-full blur-[100px] animate-float" />

      <div className="relative max-w-6xl mx-auto px-5 sm:px-6">
        <FadeIn
          y={40}
          className="relative rounded-[2rem] p-8 sm:p-14 lg:p-20 text-center overflow-hidden bg-gradient-to-br from-ink-950 via-ink-900 to-brand-900 shadow-[0_32px_100px_rgba(15,23,42,0.25)] border border-white/10"
        >
          <div className="absolute inset-0 bg-light-grid opacity-10" />
          <div className="absolute -top-32 -right-20 w-80 h-80 rounded-full bg-brand-500/25 blur-3xl" />
          <div className="absolute -bottom-32 -left-20 w-80 h-80 rounded-full bg-cyan-500/20 blur-3xl" />
          <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-brand-400 via-cyan-400 to-orange-400" />

          <div className="relative inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-white/10 border border-white/15 mb-6 shadow-lg">
            <Sparkles className="w-7 h-7 text-white" />
          </div>

          <h2 className="relative text-3xl sm:text-4xl lg:text-5xl font-black text-white leading-tight tracking-[-0.035em]">
            {content.title}
            <br />
            <span className="gradient-brand-text">{content.titleHighlight}</span>
          </h2>

          <p className="relative mt-5 text-base sm:text-lg text-ink-300 max-w-2xl mx-auto leading-relaxed">
            {content.description}
          </p>

          <div className="relative mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-3 text-sm text-ink-200">
            {content.benefits.map((benefit) => (
              <div key={benefit} className="flex items-center gap-1.5">
                <Check className="w-4 h-4 text-cyan-400" />
                {benefit}
              </div>
            ))}
          </div>

          <div className="relative mt-10 flex flex-col sm:flex-row items-center justify-center gap-3">
            <a
              href={content.primaryCta.href}
              className="group w-full sm:w-auto flex items-center justify-center gap-2 px-8 py-4 rounded-full bg-white text-ink-950 font-semibold text-base hover:bg-brand-50 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-xl"
            >
              {content.primaryCta.label}
              <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
            </a>
            <a
              href={content.secondaryCta.href}
              className="w-full sm:w-auto flex items-center justify-center gap-2 px-8 py-4 rounded-full bg-white/10 text-white font-semibold text-base border border-white/20 hover:bg-white/15 hover:border-white/30 transition-all duration-300"
            >
              {content.secondaryCta.label}
            </a>
          </div>
        </FadeIn>
      </div>
    </section>
  )
}
