import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from 'react'
import { cn } from '@/shared/cn'

interface FadeInProps {
  children: ReactNode
  className?: string
  delay?: number
  y?: number
  x?: number
  duration?: number
  margin?: string
}

export function FadeIn({
  children,
  className,
  delay = 0,
  y = 30,
  x = 0,
  duration = 0.6,
  margin = '-50px',
}: FadeInProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry?.isIntersecting) return
        setVisible(true)
        observer.disconnect()
      },
      { rootMargin: margin },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [margin])

  return (
    <div
      ref={ref}
      className={cn('fade-in', visible && 'fade-in--visible', className)}
      style={
        {
          '--fade-delay': `${delay}s`,
          '--fade-duration': `${duration}s`,
          '--fade-x': `${x}px`,
          '--fade-y': `${y}px`,
        } as CSSProperties
      }
    >
      {children}
    </div>
  )
}
