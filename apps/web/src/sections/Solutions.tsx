import { ArrowUpRight } from 'lucide-react'
import { resolveLucideIcon } from '@/shared/icons'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import type { SolutionsContent } from '@/content/types'
import { getAccent } from '@/shared/accents'

export interface SolutionsProps {
  content: SolutionsContent
}

export function Solutions({ content }: SolutionsProps) {
  return (
    <Section
      background={
        <div className="absolute top-1/3 right-0 w-[500px] h-[500px] bg-brand-200/30 rounded-full blur-[120px]" />
      }
      id="solutions"
    >
      <SectionHeader
        align="center"
        badge={content.badge}
        title={content.title}
        titleHighlight={content.titleHighlight}
        subtitle={content.subtitle}
        className="mb-20 [&_h2]:text-2xl sm:[&_h2]:text-3xl lg:[&_h2]:text-4xl [&_h2]:font-bold [&_h2]:text-ink-800 [&_h2]:tracking-[-0.025em] [&_h2_span]:inline-block [&_h2_span]:mt-6 sm:[&_h2_span]:mt-7 [&_h2_span]:text-[1.3em] [&_h2_span]:font-black [&_h2_span]:leading-none [&_h2_span]:tracking-[-0.045em]"
      />

      <div className="grid md:grid-cols-3 gap-5 lg:gap-6">
        {content.items.map((solution, index) => {
          const colors = getAccent(solution.accent)
          const Icon = resolveLucideIcon(solution.icon)
          return (
            <FadeIn
              key={solution.title}
              delay={index * 0.1}
              y={40}
              className={`group relative glass-card glass-card-hover rounded-2xl p-7 lg:p-8 ${colors.hoverBg}`}
            >
              <div
                className={`w-14 h-14 rounded-xl ${colors.iconBgSolid} ${colors.border} border flex items-center justify-center mb-6 group-hover:scale-110 transition-transform duration-300`}
              >
                <Icon className={`w-7 h-7 ${colors.iconColor}`} />
              </div>
              <h3 className="text-xl font-bold text-ink-950 mb-3">{solution.title}</h3>
              <p className="text-sm text-ink-600 leading-relaxed mb-5">{solution.description}</p>
              <ul className="space-y-2">
                {solution.points.map((point) => (
                  <li key={point} className="flex items-center gap-2 text-sm text-ink-700">
                    <div className={`w-1.5 h-1.5 rounded-full ${colors.dot}`} />
                    {point}
                  </li>
                ))}
              </ul>
              <div className="mt-6 pt-5 border-t border-ink-200/60">
                <a
                  href={solution.href}
                  className={`inline-flex items-center gap-1 text-sm font-medium ${colors.iconColor} hover:gap-2 transition-all`}
                >
                  {content.learnMoreLabel}
                  <ArrowUpRight className="w-4 h-4" />
                </a>
              </div>
            </FadeIn>
          )
        })}
      </div>
    </Section>
  )
}
