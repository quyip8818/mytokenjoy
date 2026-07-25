import { Check } from 'lucide-react'
import { resolveLucideIcon } from '@/shared/icons'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import type { CapabilitiesContent } from '@/content/types'
import { getAccent } from '@/shared/accents'

export interface CapabilitiesProps {
  content: CapabilitiesContent
}

export function Capabilities({ content }: CapabilitiesProps) {
  return (
    <Section id="capabilities" className="overflow-visible">
      <SectionHeader title={content.title} />

      <div className="grid md:grid-cols-2 gap-5 lg:gap-6">
        {content.items.map((cap, index) => {
          const accent = getAccent(cap.accent)
          const Icon = resolveLucideIcon(cap.icon)
          return (
            <FadeIn
              key={cap.title}
              delay={(index % 2) * 0.1}
              className="relative glass-card glass-card-hover rounded-2xl overflow-hidden flex flex-col"
            >
              <div className={`h-1 w-full bg-gradient-to-r ${accent.line}`} />
              <div className="p-7 lg:p-8 flex-1 flex flex-col">
                <div
                  className={`w-12 h-12 rounded-xl bg-gradient-to-br ${accent.iconBg} border border-ink-200/40 flex items-center justify-center mb-5`}
                >
                  <Icon className={`w-6 h-6 ${accent.iconColor}`} />
                </div>
                <h3 className="text-xl font-bold text-ink-950 mb-3">{cap.title}</h3>
                <p className="text-sm text-ink-600 leading-relaxed mb-6">{cap.description}</p>
                <ul className="space-y-3 mt-auto">
                  {cap.points.map((point) => (
                    <li key={point} className="flex gap-2.5">
                      <Check
                        className={`w-4 h-4 ${accent.checkColor} shrink-0 mt-0.5`}
                        strokeWidth={3}
                      />
                      <span className="text-sm text-ink-700 leading-relaxed">{point}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </FadeIn>
          )
        })}
      </div>
    </Section>
  )
}
