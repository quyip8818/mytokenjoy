import { ArrowRight, Code2 } from 'lucide-react'
import type { ReactNode } from 'react'
import { resolveLucideIcon } from '@/shared/icons'
import { Section } from '@/shared/Section'
import { SectionHeader } from '@/shared/SectionHeader'
import { FadeIn } from '@/shared/FadeIn'
import type { IntegrationContent } from '@/content/types'
import { cn } from '@/shared/cn'
import { getAccent } from '@/shared/accents'

export interface ApiCompatibilityProps {
  content: IntegrationContent
}

function CodePanel({
  label,
  highlighted,
  children,
}: {
  label: string
  highlighted?: boolean
  children: ReactNode
}) {
  return (
    <div
      className={cn(
        'relative rounded-2xl overflow-hidden bg-ink-900 shadow-2xl',
        highlighted && 'ring-2 ring-brand-500/40',
      )}
    >
      <div className="px-5 py-3 border-b border-ink-700 flex items-center gap-2">
        <div className="w-2 h-2 rounded-full bg-red-500/70" />
        <div className="w-2 h-2 rounded-full bg-yellow-500/70" />
        <div className="w-2 h-2 rounded-full bg-green-500/70" />
        <div
          className={cn(
            'ml-3 flex items-center gap-2 text-xs',
            highlighted ? 'text-brand-300' : 'text-ink-400',
          )}
        >
          <Code2 className="w-3.5 h-3.5" />
          <span className={highlighted ? 'font-medium' : undefined}>{label}</span>
        </div>
      </div>
      <pre className="px-5 py-5 text-sm font-mono text-ink-200 overflow-x-auto leading-relaxed">
        <code>{children}</code>
      </pre>
    </div>
  )
}

export function ApiCompatibility({ content }: ApiCompatibilityProps) {
  return (
    <Section id="integration" className="overflow-visible">
      <SectionHeader title={content.title} />

      <div className="grid lg:grid-cols-[1fr_auto_1fr] gap-6 items-center mb-12">
        <FadeIn x={-30} y={0} duration={0.7}>
          <CodePanel label={content.before.label}>
            <span className="text-pink-400">from</span>{' '}
            <span className="text-cyan-300">openai</span>{' '}
            <span className="text-pink-400">import</span>{' '}
            <span className="text-cyan-300">OpenAI</span>{' '}
            <span className="text-ink-500"># 直接使用厂商的 API Key</span>
            {'\n'}
            <span className="text-ink-300">client</span> ={' '}
            <span className="text-cyan-300">OpenAI</span>(
            <span className="text-ink-300">api_key</span>=
            <span className="text-amber-300">"sk-proj-original-key..."</span>){'\n'}
            <span className="text-ink-300">response</span> ={' '}
            <span className="text-ink-300">client</span>.<span className="text-cyan-300">chat</span>
            .<span className="text-cyan-300">completions</span>.
            <span className="text-cyan-300">create</span>({'\n'}
            {'  '}
            <span className="text-ink-300">model</span>=
            <span className="text-amber-300">"gpt-4"</span>,{' '}
            <span className="text-ink-300">messages</span>=[ ... ] ){'\n'}
          </CodePanel>
        </FadeIn>

        <FadeIn delay={0.4} y={0} className="hidden lg:flex justify-center">
          <div className="w-12 h-12 rounded-full bg-gradient-to-br from-brand-400 to-brand-600 flex items-center justify-center shadow-lg">
            <ArrowRight className="w-5 h-5 text-white" />
          </div>
        </FadeIn>

        <FadeIn x={30} y={0} delay={0.1} duration={0.7}>
          <CodePanel label={content.after.label} highlighted={content.after.highlighted}>
            <span className="text-pink-400">from</span>{' '}
            <span className="text-cyan-300">openai</span>{' '}
            <span className="text-pink-400">import</span>{' '}
            <span className="text-cyan-300">OpenAI</span>{' '}
            <span className="text-ink-500"># 仅需替换 Base URL 和</span>
            {'\n'}
            {'  '}
            <span className="text-ink-500">Tokenjoy 颁发的 Key</span>
            {'\n'}
            <span className="text-ink-300">client</span> ={' '}
            <span className="text-cyan-300">OpenAI</span>(
            <span className="text-ink-300">api_key</span>=
            <span className="text-amber-300">"tj-team-a-developer-key..."</span>,{'\n'}
            {'  '}
            <span className="text-ink-300">base_url</span>=
            <span className="text-amber-300">"https://app.tokenjoy.me/v1"</span>){'\n'}
            <span className="text-ink-300">response</span> ={' '}
            <span className="text-ink-300">client</span>.<span className="text-cyan-300">chat</span>
            .<span className="text-cyan-300">completions</span>.
            <span className="text-cyan-300">create</span>({'\n'}
            {'  '}
            <span className="text-ink-300">model</span>=
            <span className="text-amber-300">"gpt-4"</span>,{' '}
            <span className="text-ink-300">messages</span>=[ ... ] ){'\n'}
          </CodePanel>
        </FadeIn>
      </div>

      <div className="grid md:grid-cols-3 gap-6">
        {content.features.map((feature, index) => {
          const accent = getAccent(feature.accent)
          const Icon = resolveLucideIcon(feature.icon)
          return (
            <FadeIn
              key={feature.title}
              delay={index * 0.1}
              duration={0.5}
              className="glass-card glass-card-hover rounded-2xl p-7"
            >
              <div
                className={`w-11 h-11 rounded-xl bg-gradient-to-br ${accent.iconBg} border border-ink-200/40 flex items-center justify-center mb-4`}
              >
                <Icon className={`w-5 h-5 ${accent.iconColor}`} />
              </div>
              <h3 className="text-lg font-bold text-ink-950 mb-3">{feature.title}</h3>
              <p className="text-sm text-ink-600 leading-relaxed">{feature.description}</p>
            </FadeIn>
          )
        })}
      </div>
    </Section>
  )
}
