import { resolveLucideIcon } from '@/shared/icons'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import { RichText } from '@/shared/RichText'
import type { MissionContent } from '@/content/types'

export interface MissionProps {
  content: MissionContent
}

export function Mission({ content }: MissionProps) {
  return (
    <Section
      id="mission"
      background={
        <>
          <div className="absolute inset-0 bg-light-grid opacity-30" />
          <div className="absolute top-1/3 right-0 w-[500px] h-[500px] bg-brand-200/30 rounded-full blur-[120px]" />
        </>
      }
    >
      <SectionHeader title={content.title} />

      <FadeIn
        delay={0.1}
        className="relative rounded-2xl bg-gradient-to-br from-brand-50/80 via-white to-cyan-50/40 border border-brand-100/60 p-8 lg:p-12 mb-10"
      >
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-brand-300 to-transparent" />
        <p className="text-center text-xl sm:text-2xl lg:text-3xl font-semibold gradient-brand-text leading-relaxed">
          "{content.quote}"
        </p>
        <p className="mt-6 text-center text-sm sm:text-base text-ink-600 leading-relaxed max-w-3xl mx-auto">
          <RichText text={content.description} />
        </p>
      </FadeIn>

      <div className="grid md:grid-cols-2 gap-6">
        {content.cards.map((card, index) => {
          const Icon = resolveLucideIcon(card.icon)
          return (
            <FadeIn
              key={card.title}
              delay={index * 0.1}
              className="glass-card glass-card-hover rounded-2xl p-8"
            >
              <div className="flex items-center gap-3 mb-5">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-100 to-brand-50 border border-brand-200/60 flex items-center justify-center">
                  <Icon className="w-5 h-5 text-brand-600" />
                </div>
                <h3 className="text-xl font-bold text-ink-950">{card.title}</h3>
              </div>
              <p className="text-sm text-ink-600 leading-relaxed">{card.content}</p>
            </FadeIn>
          )
        })}
      </div>
    </Section>
  )
}
