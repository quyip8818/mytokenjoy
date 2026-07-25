import type { ComponentType } from 'react'
import {
  Activity,
  CheckCircle2,
  Cloud,
  Eye,
  GitBranch,
  Layers,
  MessageCircle,
  Plug,
  Server,
  ShieldCheck,
  Users,
  Wallet,
} from 'lucide-react'
import {
  ChallengeBudgetIcon,
  ChallengeObservabilityIcon,
  ChallengeSecurityIcon,
} from '@/shared/ChallengeIcons'
import { GithubIcon, LinkedinIcon, TwitterIcon } from '@/shared/BrandIcons'

const lucideIcons = {
  wallet: Wallet,
  shield: ShieldCheck,
  activity: Activity,
  plug: Plug,
  check: CheckCircle2,
  branch: GitBranch,
  layers: Layers,
  cloud: Cloud,
  server: Server,
  eye: Eye,
  users: Users,
  github: GithubIcon,
  twitter: TwitterIcon,
  linkedin: LinkedinIcon,
  wechat: MessageCircle,
} as const

export const challengeIcons = {
  budget: ChallengeBudgetIcon,
  security: ChallengeSecurityIcon,
  observability: ChallengeObservabilityIcon,
} as const

export function resolveLucideIcon(
  key: keyof typeof lucideIcons,
): ComponentType<{ className?: string }> {
  return lucideIcons[key]
}
