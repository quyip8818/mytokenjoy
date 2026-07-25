import { Check } from 'lucide-react'
import { resolveLucideIcon } from '@/shared/icons'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import type { DeploymentContent } from '@/content/types'
import { getAccent } from '@/shared/accents'

export interface DeploymentModesProps {
  content: DeploymentContent
}

export function DeploymentModes({ content }: DeploymentModesProps) {
  return (
    <Section
      id="deployment"
      background={<div className="absolute inset-0 bg-light-grid opacity-30" />}
    >
      <SectionHeader title={content.title} className="mb-16" />

      <div className="grid lg:grid-cols-2 gap-5 lg:gap-6">
        {content.modes.map((mode, index) => {
          const accent = getAccent(mode.accent)
          const Icon = resolveLucideIcon(mode.icon)
          return (
            <FadeIn
              key={mode.id}
              delay={index * 0.1}
              y={40}
              className="relative glass-card glass-card-hover rounded-2xl overflow-hidden flex flex-col"
            >
              <div className={`h-1 w-full bg-gradient-to-r ${accent.line}`} />
              <div className="p-7 sm:p-8 lg:p-10 flex-1 flex flex-col">
                <div
                  className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${accent.iconBg} border ${accent.tagBorder} flex items-center justify-center mb-5`}
                >
                  <Icon className={`w-7 h-7 ${accent.iconColor}`} />
                </div>
                <h3 className="text-2xl font-bold text-ink-950 mb-2">{mode.title}</h3>
                <p className="text-sm text-ink-500 mb-8">{mode.subtitle}</p>
                <ul className="space-y-5 flex-1">
                  {mode.points.map((point) => (
                    <li key={point.label} className="flex gap-3">
                      <Check
                        className={`w-5 h-5 ${accent.checkColor} shrink-0 mt-0.5`}
                        strokeWidth={2.5}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-semibold text-ink-950 text-sm">{point.label}</span>
                          {point.tag ? (
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded ${accent.tagBg} ${accent.tagText} ${accent.tagBorder} border text-xs font-medium`}
                            >
                              {point.tag}
                            </span>
                          ) : null}
                        </div>
                        <p className="text-sm text-ink-600 leading-relaxed mt-1">{point.desc}</p>
                      </div>
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
