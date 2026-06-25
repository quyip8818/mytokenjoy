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
import { PERMISSION, type PermissionKey, hasPermission } from '@/lib/permissions'

export interface NavItem {
  label: string
  path: string
  icon: LucideIcon
  requiredPermissions: PermissionKey[]
  badgeKey?: 'approvalPending'
}

export interface NavGroup {
  group: string
  items: NavItem[]
  collapsed?: boolean
}

export const NAV_GROUPS: NavGroup[] = [
  {
    group: '组织',
    items: [
      {
        label: '数据源',
        path: '/org/data-source',
        icon: Database,
        requiredPermissions: [PERMISSION.ORG_DATASOURCE],
      },
      {
        label: '组织架构',
        path: '/org/structure',
        icon: Building2,
        requiredPermissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS],
      },
      {
        label: '角色管理',
        path: '/org/roles',
        icon: Shield,
        requiredPermissions: [PERMISSION.ORG_ROLES],
      },
    ],
  },
  {
    group: '预算',
    items: [
      {
        label: '预算总览',
        path: '/budget/overview',
        icon: Wallet,
        requiredPermissions: [PERMISSION.BUDGET_READ],
      },
      {
        label: '预算分配',
        path: '/budget/allocation',
        icon: PieChart,
        requiredPermissions: [PERMISSION.BUDGET_READ],
      },
      {
        label: '超限策略',
        path: '/budget/alerts',
        icon: ShieldAlert,
        requiredPermissions: [PERMISSION.BUDGET_POLICY],
      },
    ],
  },
  {
    group: '模型管理',
    items: [
      {
        label: '模型列表',
        path: '/models/list',
        icon: Cpu,
        requiredPermissions: [PERMISSION.MODEL_MANAGE],
      },
      {
        label: '模型白名单',
        path: '/models/routing',
        icon: GitBranch,
        requiredPermissions: [PERMISSION.MODEL_WHITELIST],
      },
    ],
  },
  {
    group: 'Key 中心',
    items: [
      {
        label: '我的 Key',
        path: '/keys/mine',
        icon: CreditCard,
        requiredPermissions: [PERMISSION.SELF_KEYS],
      },
      {
        label: '审批中心',
        path: '/keys/approval',
        icon: CheckCircle2,
        requiredPermissions: [PERMISSION.BUDGET_APPROVE, PERMISSION.SELF_APPROVAL],
        badgeKey: 'approvalPending',
      },
      {
        label: 'Key 管理',
        path: '/keys/platform',
        icon: Globe,
        requiredPermissions: [PERMISSION.KEYS_ADMIN],
      },
      {
        label: '供应商 Key',
        path: '/keys/provider',
        icon: Key,
        requiredPermissions: [PERMISSION.KEYS_PROVIDER],
      },
    ],
  },
  {
    group: '数据中心',
    collapsed: true,
    items: [
      {
        label: '成本看板',
        path: '/dashboard/cost',
        icon: BarChart3,
        requiredPermissions: [PERMISSION.DASHBOARD_COST],
      },
      {
        label: '用量分析',
        path: '/dashboard/usage',
        icon: TrendingUp,
        requiredPermissions: [PERMISSION.DASHBOARD_USAGE],
      },
    ],
  },
  {
    group: '审计',
    collapsed: true,
    items: [
      {
        label: '操作审计',
        path: '/audit/operations',
        icon: ScrollText,
        requiredPermissions: [PERMISSION.AUDIT_READ],
      },
      {
        label: '调用日志',
        path: '/audit/calls',
        icon: Activity,
        requiredPermissions: [PERMISSION.AUDIT_READ],
      },
    ],
  },
]

export const ROUTE_TITLES: Record<string, string> = Object.fromEntries(
  NAV_GROUPS.flatMap((group) => group.items.map((item) => [item.path, item.label])),
)

export function getVisibleNavGroups(permissions: readonly string[]): NavGroup[] {
  return NAV_GROUPS.map((group) => ({
    ...group,
    collapsed: group.collapsed && !group.items.some((item) => hasNavPermission(permissions, item)),
    items: group.items.filter((item) => hasNavPermission(permissions, item)),
  })).filter((group) => group.items.length > 0)
}

function hasNavPermission(permissions: readonly string[], item: NavItem): boolean {
  return hasPermission(permissions, item.requiredPermissions)
}
