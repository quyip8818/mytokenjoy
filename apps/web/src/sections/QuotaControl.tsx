import { ArrowDown, CheckCircle2, Users } from 'lucide-react'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import { RichText } from '@/shared/RichText'
import type { QuotaContent } from '@/content/types'
import { getAccent } from '@/shared/accents'

export interface QuotaControlProps {
  content: QuotaContent
}

export function QuotaControl({ content }: QuotaControlProps) {
  return (
    <Section
      id="quota"
      className="bg-ink-100/30"
      background={<div className="absolute inset-0 bg-dot-pattern opacity-30" />}
    >
      <SectionHeader title={content.title} />

      <div className="grid lg:grid-cols-2 gap-12">
        <FadeIn x={-30} y={0} duration={0.7}>
          <div className="flex items-center gap-2 mb-6">
            <Users className="w-5 h-5 text-brand-600" />
            <h3 className="text-lg font-bold text-ink-950">{content.flowTitle}</h3>
          </div>
          <div className="space-y-3">
            {content.levels.map((level, index) => (
              <FadeIn key={level.title} delay={index * 0.1} y={20} duration={0.4}>
                <div
                  className="relative rounded-2xl p-5 shadow-lg"
                  style={{ background: level.background }}
                >
                  <div className="text-center">
                    <div className="text-lg font-bold text-white">{level.title}</div>
                    <div className={`text-xs mt-1 ${level.subtitleClass}`}>{level.subtitle}</div>
                  </div>
                </div>
                {index < content.levels.length - 1 ? (
                  <div className="flex justify-center py-2 text-ink-400">
                    <ArrowDown className="w-5 h-5" />
                  </div>
                ) : null}
              </FadeIn>
            ))}
          </div>
        </FadeIn>

        <FadeIn x={30} y={0} delay={0.1} duration={0.7}>
          <div className="flex items-center gap-2 mb-6">
            <CheckCircle2 className="w-5 h-5 text-brand-600" />
            <h3 className="text-lg font-bold text-ink-950">{content.featuresTitle}</h3>
          </div>
          <div className="space-y-4">
            {content.features.map((feature, index) => {
              const accent = getAccent(feature.accent)
              return (
                <FadeIn
                  key={feature.title}
                  delay={0.1 + index * 0.1}
                  y={20}
                  duration={0.4}
                  className={`bg-white rounded-xl p-6 border-l-4 ${accent.borderLeft} shadow-sm hover:shadow-md transition-shadow`}
                >
                  <h4 className="text-base font-bold text-ink-950 mb-2">{feature.title}</h4>
                  <p className="text-sm text-ink-600 leading-relaxed">
                    <RichText
                      text={feature.description}
                      boldClassName="text-brand-600 font-semibold"
                    />
                  </p>
                </FadeIn>
              )
            })}
          </div>
        </FadeIn>
      </div>
    </Section>
  )
}
