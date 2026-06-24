import {
  Activity,
  BarChart3,
  Building2,
  CheckCircle2,
  Cpu,
  CreditCard,
  Database,
  GitBranch,
  Globe,
  Key,
  PieChart,
  ScrollText,
  Shield,
  ShieldAlert,
  TrendingUp,
  Wallet,
  type LucideIcon,
} from 'lucide-react'
import type { DemoRole } from '@/features/demo'

export interface NavItem {
  label: string
  path: string
  icon: LucideIcon
  roles: DemoRole[]
  badgeKey?: 'approvalPending'
}

export interface NavGroup {
  group: string
  items: NavItem[]
  roles?: DemoRole[]
  collapsed?: boolean
}

export const NAV_GROUPS: NavGroup[] = [
  {
    group: '组织',
    items: [
      { label: '数据源', path: '/org/data-source', icon: Database, roles: ['admin'] },
      { label: '组织架构', path: '/org/structure', icon: Building2, roles: ['admin'] },
      { label: '角色管理', path: '/org/roles', icon: Shield, roles: ['admin'] },
    ],
  },
  {
    group: '预算',
    items: [
      { label: '预算总览', path: '/budget/overview', icon: Wallet, roles: ['admin', 'tl'] },
      { label: '预算分配', path: '/budget/allocation', icon: PieChart, roles: ['admin', 'tl'] },
      { label: '超限策略', path: '/budget/alerts', icon: ShieldAlert, roles: ['admin'] },
    ],
  },
  {
    group: '模型管理',
    items: [
      { label: '模型列表', path: '/models/list', icon: Cpu, roles: ['admin', 'tl'] },
      { label: '模型白名单', path: '/models/routing', icon: GitBranch, roles: ['admin', 'tl'] },
    ],
  },
  {
    group: 'Key 中心',
    items: [
      { label: '我的 Key', path: '/keys/mine', icon: CreditCard, roles: ['member', 'tl', 'admin'] },
      {
        label: '审批中心',
        path: '/keys/approval',
        icon: CheckCircle2,
        roles: ['admin', 'tl', 'member'],
        badgeKey: 'approvalPending',
      },
      { label: 'Key 管理', path: '/keys/platform', icon: Globe, roles: ['admin'] },
      { label: '供应商 Key', path: '/keys/provider', icon: Key, roles: ['admin'] },
    ],
  },
  {
    group: '数据中心',
    collapsed: true,
    items: [
      { label: '成本看板', path: '/dashboard/cost', icon: BarChart3, roles: ['admin'] },
      { label: '用量分析', path: '/dashboard/usage', icon: TrendingUp, roles: ['admin'] },
    ],
  },
  {
    group: '审计',
    collapsed: true,
    items: [
      { label: '操作审计', path: '/audit/operations', icon: ScrollText, roles: ['admin'] },
      { label: '调用日志', path: '/audit/calls', icon: Activity, roles: ['admin'] },
    ],
  },
]

export const ROUTE_TITLES: Record<string, string> = Object.fromEntries(
  NAV_GROUPS.flatMap((group) => group.items.map((item) => [item.path, item.label])),
)

export function getVisibleNavGroups(role: DemoRole): NavGroup[] {
  return NAV_GROUPS.map((group) => ({
    ...group,
    items: group.items.filter((item) => item.roles.includes(role)),
  })).filter((group) => group.items.length > 0)
}
