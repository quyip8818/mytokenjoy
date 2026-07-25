import { Section } from '@/shared/Section'
import { FadeIn } from '@/shared/FadeIn'
import type { TestimonialsContent } from '@/content/types'

export interface TestimonialsProps {
  content: TestimonialsContent
}

export function Testimonials({ content }: TestimonialsProps) {
  return (
    <Section
      id="testimonials"
      background={
        <div className="absolute top-1/4 left-0 w-[400px] h-[400px] bg-cyan-200/20 rounded-full blur-[120px]" />
      }
    >
      <FadeIn className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4 mb-16 sm:mb-20">
        {content.stats.map((stat, index) => (
          <FadeIn
            key={stat.label}
            delay={index * 0.1}
            y={20}
            className="glass-card rounded-2xl p-5 sm:p-6 text-center"
          >
            <div className="text-3xl sm:text-4xl font-black gradient-brand-text tracking-tight">
              {stat.value}
            </div>
            <div className="mt-2 text-sm text-ink-600">{stat.label}</div>
          </FadeIn>
        ))}
      </FadeIn>

      <FadeIn className="text-center max-w-3xl mx-auto mb-14 sm:mb-16">
        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-black text-ink-950 leading-tight tracking-[-0.035em]">
          {content.title}
          <span className="gradient-brand-text">{content.titleHighlight}</span>
        </h2>
        <p className="mt-5 text-base text-ink-600">{content.subtitle}</p>
      </FadeIn>

      <div className="grid md:grid-cols-3 gap-6">
        {content.items.map((item, index) => (
          <FadeIn
            key={item.name}
            delay={index * 0.1}
            y={40}
            className="glass-card glass-card-hover rounded-2xl p-7 flex flex-col"
          >
            <div className="text-6xl h-9 leading-none gradient-brand-text font-serif">“</div>
            <p className="text-sm text-ink-700 leading-relaxed flex-1 mt-5">{item.quote}</p>
            <div className="mt-6 flex items-center gap-3 pt-5 border-t border-ink-200/60">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-400 to-brand-600 flex items-center justify-center">
                <span className="text-white font-semibold text-sm">{item.name.charAt(0)}</span>
              </div>
              <div>
                <div className="font-medium text-ink-950 text-sm">{item.name}</div>
                <div className="text-xs text-ink-500">{item.role}</div>
              </div>
            </div>
          </FadeIn>
        ))}
      </div>
    </Section>
  )
}
