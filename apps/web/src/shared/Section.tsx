import type { ReactNode } from 'react'
import { cn } from '@/shared/cn'

interface SectionProps {
  id?: string
  children: ReactNode
  className?: string
  background?: ReactNode
}

export function Section({ id, children, className, background }: SectionProps) {
  return (
    <section
      id={id}
      className={cn('relative py-20 sm:py-24 lg:py-32 overflow-hidden scroll-mt-20', className)}
    >
      {background}
      <div className="relative max-w-7xl mx-auto px-5 sm:px-6 lg:px-8">{children}</div>
    </section>
  )
}
