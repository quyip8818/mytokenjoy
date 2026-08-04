import type { LucideIcon } from 'lucide-react'
import type { AppSession } from '@/features/session/types'

/** A single feature card shown inside an announcement dialog. */
export interface AnnouncementFeature {
  icon: LucideIcon
  title: string
  description: string
  /** Tailwind color classes for card accent, e.g. 'border-indigo-100 bg-indigo-50/50' */
  cardClass: string
  /** Tailwind color classes for icon container, e.g. 'bg-indigo-100 text-indigo-600' */
  iconClass: string
}

/** Declarative announcement configuration. */
export interface AnnouncementConfig {
  /** Unique ID — used as localStorage key suffix. */
  id: string
  /** Header gradient classes. */
  gradient: string
  /** Header icon. */
  icon: LucideIcon
  /** Header title. */
  title: string
  /** Header subtitle. */
  subtitle: string
  /** Feature cards. */
  features: AnnouncementFeature[]
  /** CTA button label. */
  ctaLabel: string
  /** Show condition: return true to display this announcement. */
  shouldShow: (session: AppSession) => boolean
}
