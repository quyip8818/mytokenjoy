import { useEffect, useState, type ComponentType } from 'react'
import { Hero } from '@/sections/Hero'
import { Navbar } from '@/shared'
import { homeContent, type HomeContent } from '@/content'
import type { HomeRestProps } from '@/pages/HomeRest'

export interface HomeProps {
  content?: HomeContent
}

export default function Home({ content = homeContent }: HomeProps) {
  return (
    <div className="relative min-h-screen bg-white text-ink-950 overflow-x-hidden">
      <Navbar content={content.nav} />
      <main>
        <Hero content={content.hero} />
        <DeferredHomeRest content={content} />
      </main>
    </div>
  )
}

function DeferredHomeRest({ content }: { content: HomeContent }) {
  const [Rest, setRest] = useState<ComponentType<HomeRestProps> | null>(null)

  useEffect(() => {
    let cancelled = false

    const load = () => {
      void import('@/pages/HomeRest').then((mod) => {
        if (!cancelled) setRest(() => mod.default)
      })
    }

    if (typeof window.requestIdleCallback === 'function') {
      const id = window.requestIdleCallback(load, { timeout: 1500 })
      return () => {
        cancelled = true
        window.cancelIdleCallback(id)
      }
    }

    const id = window.setTimeout(load, 200)
    return () => {
      cancelled = true
      window.clearTimeout(id)
    }
  }, [])

  if (!Rest) return null
  return <Rest content={content} />
}
