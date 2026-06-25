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
import { type PermissionKey, hasPermission } from '@/lib/permissions'
import { ROUTES, routePermissions } from '@/config/routes'

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
        path: ROUTES.orgDataSource,
        icon: Database,
        requiredPermissions: routePermissions(ROUTES.orgDataSource),
      },
      {
        label: '组织架构',
        path: ROUTES.orgStructure,
        icon: Building2,
        requiredPermissions: routePermissions(ROUTES.orgStructure),
      },
      {
        label: '角色管理',
        path: ROUTES.orgRoles,
        icon: Shield,
        requiredPermissions: routePermissions(ROUTES.orgRoles),
      },
    ],
  },
  {
    group: '预算',
    items: [
      {
        label: '预算总览',
        path: ROUTES.budgetOverview,
        icon: Wallet,
        requiredPermissions: routePermissions(ROUTES.budgetOverview),
      },
      {
        label: '预算分配',
        path: ROUTES.budgetAllocation,
        icon: PieChart,
        requiredPermissions: routePermissions(ROUTES.budgetAllocation),
      },
      {
        label: '超限策略',
        path: ROUTES.budgetAlerts,
        icon: ShieldAlert,
        requiredPermissions: routePermissions(ROUTES.budgetAlerts),
      },
    ],
  },
  {
    group: '模型管理',
    items: [
      {
        label: '模型列表',
        path: ROUTES.modelsList,
        icon: Cpu,
        requiredPermissions: routePermissions(ROUTES.modelsList),
      },
      {
        label: '模型白名单',
        path: ROUTES.modelsRouting,
        icon: GitBranch,
        requiredPermissions: routePermissions(ROUTES.modelsRouting),
      },
    ],
  },
  {
    group: 'Key 中心',
    items: [
      {
        label: '我的 Key',
        path: ROUTES.keysMine,
        icon: CreditCard,
        requiredPermissions: routePermissions(ROUTES.keysMine),
      },
      {
        label: '审批中心',
        path: ROUTES.keysApproval,
        icon: CheckCircle2,
        requiredPermissions: routePermissions(ROUTES.keysApproval),
        badgeKey: 'approvalPending',
      },
      {
        label: 'Key 管理',
        path: ROUTES.keysPlatform,
        icon: Globe,
        requiredPermissions: routePermissions(ROUTES.keysPlatform),
      },
      {
        label: '供应商 Key',
        path: ROUTES.keysProvider,
        icon: Key,
        requiredPermissions: routePermissions(ROUTES.keysProvider),
      },
    ],
  },
  {
    group: '数据中心',
    collapsed: true,
    items: [
      {
        label: '成本看板',
        path: ROUTES.dashboardCost,
        icon: BarChart3,
        requiredPermissions: routePermissions(ROUTES.dashboardCost),
      },
      {
        label: '用量分析',
        path: ROUTES.dashboardUsage,
        icon: TrendingUp,
        requiredPermissions: routePermissions(ROUTES.dashboardUsage),
      },
    ],
  },
  {
    group: '审计',
    collapsed: true,
    items: [
      {
        label: '操作审计',
        path: ROUTES.auditOperations,
        icon: ScrollText,
        requiredPermissions: routePermissions(ROUTES.auditOperations),
      },
      {
        label: '调用日志',
        path: ROUTES.auditCalls,
        icon: Activity,
        requiredPermissions: routePermissions(ROUTES.auditCalls),
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
