import type { AccentKey } from '@/shared/accents'
import type { ChallengesContent } from '@/content/types'
import { challengeIcons } from '@/shared/icons'
import { FadeIn } from '@/shared/FadeIn'
import { getAccent } from '@/shared/accents'
import { RichText } from '@/shared/RichText'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'

const CHALLENGE_ICON_BG: Record<'brand' | 'cyan' | 'orange', string> = {
  brand: 'from-brand-100 to-brand-50',
  cyan: 'from-cyan-50 to-white',
  orange: 'from-orange-50 to-white',
}

export interface ChallengesProps {
  content: ChallengesContent
}

export function Challenges({ content }: ChallengesProps) {
  return (
    <Section
      id="challenges"
      background={<div className="absolute inset-0 bg-light-grid opacity-30" />}
    >
      <SectionHeader title={content.title} subtitle={content.subtitle} className="mb-16" />

      <div className="grid md:grid-cols-3 gap-6">
        {content.items.map((challenge, index) => {
          const colors = getAccent(challenge.accent)
          const Icon = challengeIcons[challenge.icon]
          const iconBg =
            CHALLENGE_ICON_BG[challenge.accent as Exclude<AccentKey, 'ink'>] ?? colors.iconBg
          return (
            <FadeIn
              key={challenge.id}
              delay={index * 0.1}
              y={40}
              className="relative glass-card glass-card-hover rounded-2xl overflow-hidden"
            >
              <div className={`h-1 w-full bg-gradient-to-r ${colors.line}`} />
              <div className="p-8">
                <div
                  className={`w-16 h-16 rounded-xl bg-gradient-to-br ${iconBg} flex items-center justify-center mb-6 border border-white`}
                >
                  <Icon />
                </div>
                <div className="flex items-baseline gap-2 mb-4">
                  <span className={`text-sm font-bold ${colors.numText} font-mono`}>
                    {challenge.id}
                  </span>
                  <span className="text-ink-500">/</span>
                  <h3 className="text-xl font-bold text-ink-950">{challenge.title}</h3>
                </div>
                <p className="text-sm text-ink-600 leading-relaxed">
                  <RichText text={challenge.description} />
                </p>
              </div>
            </FadeIn>
          )
        })}
      </div>
    </Section>
  )
}
