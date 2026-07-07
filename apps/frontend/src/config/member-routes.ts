import type { ComponentType } from 'react'
import { FileText, Key, LayoutDashboard, type LucideIcon } from 'lucide-react'

type LazyPageModule = { default: ComponentType }

export interface MemberRouteDefinition {
  key: string
  path: `/me` | `/me/keys` | `/me/call-logs`
  label: string
  icon: LucideIcon
  lazy: () => Promise<LazyPageModule>
  navEnd?: boolean
}

export const MEMBER_ROUTE_DEFINITIONS: MemberRouteDefinition[] = [
  {
    key: 'memberDashboard',
    path: '/me',
    label: '工作台',
    icon: LayoutDashboard,
    lazy: () => import('@/routes/member'),
    navEnd: true,
  },
  {
    key: 'memberKeys',
    path: '/me/keys',
    label: '我的 Key',
    icon: Key,
    lazy: () => import('@/routes/member/keys'),
  },
  {
    key: 'memberCallLogs',
    path: '/me/call-logs',
    label: '使用记录',
    icon: FileText,
    lazy: () => import('@/routes/member/call-logs'),
  },
]

export function toMemberRouterPath(path: MemberRouteDefinition['path']): string {
  return path.slice(1)
}

const MEMBER_LAZY_IMPORT_PATTERN = /import\(['"](@\/routes\/[^'"]+)['"]\)/

export function getMemberRouteLazyImportPaths(): string[] {
  return MEMBER_ROUTE_DEFINITIONS.map((definition) => {
    const match = definition.lazy.toString().match(MEMBER_LAZY_IMPORT_PATTERN)
    if (!match?.[1]) {
      throw new Error(`Unable to resolve lazy import for member route ${definition.key}`)
    }
    return match[1]
  })
}
