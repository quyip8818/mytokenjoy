import { Mail } from 'lucide-react'
import { resolveLucideIcon } from '@/shared/icons'
import { Logo } from '@/shared/Logo'
import type { FooterContent } from '@/content/types'

export interface FooterProps {
  content: FooterContent
}

export function Footer({ content }: FooterProps) {
  return (
    <footer className="relative bg-ink-50 border-t border-ink-200/60 pt-20 pb-10">
      <div className="max-w-7xl mx-auto px-6 lg:px-8">
        <div className="grid lg:grid-cols-6 gap-12">
          <div className="lg:col-span-2">
            <Logo size="md" />
            <p className="mt-4 text-sm text-ink-600 leading-relaxed max-w-xs">
              {content.description}
            </p>

            <div className="mt-6 flex items-center gap-2 text-sm text-ink-600">
              <Mail className="w-4 h-4" />
              <a href={`mailto:${content.email}`} className="hover:text-ink-950">
                {content.email}
              </a>
            </div>

            <div className="flex gap-3 mt-6">
              {content.socialLinks.map((social) => {
                const Icon = resolveLucideIcon(social.icon)
                return (
                  <a
                    key={social.label}
                    href={social.href}
                    aria-label={social.label}
                    className="w-9 h-9 rounded-lg bg-white border border-ink-200 flex items-center justify-center text-ink-600 hover:bg-brand-50 hover:text-brand-600 hover:border-brand-200 transition-all"
                  >
                    <Icon className="w-4 h-4" />
                  </a>
                )
              })}
            </div>
          </div>

          {content.columns.map((column) => (
            <div key={column.title}>
              <h4 className="font-semibold text-sm text-ink-950 mb-4">{column.title}</h4>
              <ul className="space-y-3">
                {column.links.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      className="text-sm text-ink-600 hover:text-ink-950 transition-colors"
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-16 pt-8 border-t border-ink-200/60 flex flex-col sm:flex-row items-center justify-center gap-2 sm:gap-4 text-center text-xs text-ink-500">
          <p>{content.copyright}</p>
          <a
            href={content.icpUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-ink-950 transition-colors"
          >
            {content.icpNumber}
          </a>
        </div>
      </div>
    </footer>
  )
}
