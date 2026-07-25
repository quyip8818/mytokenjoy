import { useState, useEffect, useRef, useCallback } from 'react'
import { ArrowRight, Menu, X } from 'lucide-react'
import type { NavContent } from '@/content/types'
import { Logo } from '@/shared/Logo'
import { cn } from '@/shared/cn'
import { useScrollThreshold } from '@/shared/useScrollThreshold'

const AUTH_EMBED_URL = import.meta.env.VITE_AUTH_EMBED_URL || 'https://app.tokenjoy.com/auth/embed'
const APP_ORIGIN = import.meta.env.VITE_APP_ORIGIN || 'https://app.tokenjoy.com'
const APP_URL = import.meta.env.VITE_APP_URL || 'https://app.tokenjoy.com'

type AuthMode = 'login' | 'register'

export interface NavbarProps {
  content: NavContent
  scrollThreshold?: number
}

export function Navbar({ content, scrollThreshold = 20 }: NavbarProps) {
  const scrolled = useScrollThreshold(scrollThreshold)
  const [mobileOpen, setMobileOpen] = useState(false)

  // --- Auth iframe dialog state ---
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<AuthMode>('login')
  const [iframeReady, setIframeReady] = useState(false)
  const [iframeError, setIframeError] = useState(false)
  const [iframeKey, setIframeKey] = useState(0)
  const dialogRef = useRef<HTMLDialogElement>(null)
  const readyRef = useRef(false)

  const openAuth = useCallback((mode: AuthMode) => {
    setAuthMode(mode)
    setAuthOpen(true)
    setIframeReady(false)
    setIframeError(false)
    setMobileOpen(false)
    readyRef.current = false
  }, [])

  const closeAuth = useCallback(() => {
    dialogRef.current?.close()
    setAuthOpen(false)
  }, [])

  useEffect(() => {
    if (!authOpen) return
    dialogRef.current?.showModal()

    const timeout = setTimeout(() => {
      if (!readyRef.current) setIframeError(true)
    }, 5000)

    const handler = (e: MessageEvent) => {
      if (e.origin !== APP_ORIGIN) return
      if (e.data?.type === 'auth:ready') {
        readyRef.current = true
        setIframeReady(true)
        setIframeError(false)
      }
      if (e.data?.type === 'auth:success') {
        closeAuth()
        window.location.href = APP_URL
      }
    }

    window.addEventListener('message', handler)
    return () => {
      clearTimeout(timeout)
      window.removeEventListener('message', handler)
    }
  }, [authOpen, closeAuth])

  return (
    <>
      <nav
        className={cn(
          'fixed top-0 left-0 right-0 z-50 transition-all duration-300 animate-fade-down px-3 sm:px-5',
          scrolled ? 'pt-2' : 'pt-3 sm:pt-4',
        )}
      >
        <div
          className={cn(
            'max-w-7xl mx-auto px-4 sm:px-5 lg:px-6 rounded-2xl transition-all duration-300',
            scrolled
              ? 'bg-white/85 backdrop-blur-2xl border border-white shadow-[0_8px_32px_rgba(15,23,42,0.09)]'
              : 'bg-white/45 backdrop-blur-md border border-white/60',
          )}
        >
          <div className="flex items-center justify-between h-16 sm:h-[68px]">
            <a href={content.homeHref} className="flex items-center">
              <Logo size="sm" />
            </a>

            <div className="hidden md:flex items-center gap-1 rounded-full bg-ink-100/60 p-1">
              {content.links.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  className="px-3.5 py-2 rounded-full text-sm text-ink-600 hover:text-ink-950 hover:bg-white transition-all font-medium"
                >
                  {link.label}
                </a>
              ))}
            </div>

            <div className="hidden md:flex items-center gap-3">
              <button
                type="button"
                onClick={() => openAuth('login')}
                className="text-sm text-ink-700 hover:text-ink-950 transition-colors font-medium"
              >
                {content.loginLabel}
              </button>
              <button
                type="button"
                onClick={() => openAuth('register')}
                className="group flex items-center gap-1.5 px-5 py-2.5 rounded-full bg-ink-950 text-white text-sm font-semibold hover:bg-brand-600 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg"
              >
                {content.ctaLabel}
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform" />
              </button>
            </div>

            <button
              type="button"
              className="md:hidden w-10 h-10 rounded-xl bg-white/80 border border-ink-200 flex items-center justify-center text-ink-950"
              onClick={() => setMobileOpen((open) => !open)}
              aria-label="Toggle menu"
              aria-expanded={mobileOpen}
            >
              {mobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>

        <div
          className={cn(
            'md:hidden grid transition-[grid-template-rows] duration-300 ease-out max-w-7xl mx-auto bg-white/95 backdrop-blur-xl rounded-b-2xl border-x border-b border-white',
            mobileOpen ? 'grid-rows-[1fr] shadow-lg' : 'grid-rows-[0fr]',
            !mobileOpen && 'border-transparent',
          )}
        >
          <div className="overflow-hidden min-h-0">
            <div className="px-4 pb-4 pt-2 space-y-1">
              {content.links.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  onClick={() => setMobileOpen(false)}
                  className="block rounded-xl px-3 py-3 text-ink-700 hover:text-ink-950 hover:bg-ink-100 font-medium transition-colors"
                >
                  {link.label}
                </a>
              ))}
              <button
                type="button"
                onClick={() => openAuth('register')}
                className="flex w-full items-center justify-center gap-1.5 px-5 py-3 rounded-full bg-ink-950 text-white font-medium"
              >
                {content.ctaLabel}
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </nav>

      <dialog
        ref={dialogRef}
        className="backdrop:bg-black/50 bg-transparent p-0 m-auto rounded-2xl outline-none"
        onClose={() => setAuthOpen(false)}
      >
        <div className="relative w-[480px] max-w-[95vw] h-[640px] max-h-[90vh] rounded-2xl overflow-hidden">
          <button
            type="button"
            onClick={closeAuth}
            className="absolute top-3 right-3 z-10 w-8 h-8 flex items-center justify-center rounded-full text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
            aria-label="关闭"
          >
            <X className="w-4 h-4" />
          </button>

          {iframeError && !iframeReady ? (
            <div className="flex h-full items-center justify-center text-center p-6 bg-white rounded-2xl shadow-2xl">
              <div>
                <p className="text-ink-600 mb-4">加载失败，请重试</p>
                <button
                  type="button"
                  onClick={() => { setIframeError(false); setIframeReady(false); setIframeKey((k) => k + 1) }}
                  className="text-brand-600 underline text-sm"
                >
                  重新加载
                </button>
              </div>
            </div>
          ) : (
            <iframe
              key={iframeKey}
              src={`${AUTH_EMBED_URL}?mode=${authMode}`}
              className="w-full h-full border-0"
              title="TokenJoy 认证"
            />
          )}
        </div>
      </dialog>
    </>
  )
}
