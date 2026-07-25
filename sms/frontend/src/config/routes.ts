import type { ComponentType } from 'react'
import {
  Box,
  FileText,
  Globe,
  LayoutDashboard,
  Scale,
  Settings,
  ShoppingCart,
  Users,
  type LucideIcon,
} from 'lucide-react'

type LazyPageModule = { default: ComponentType }

export interface RouteDefinition {
  key: string
  path: string
  label: string
  icon: LucideIcon
  lazy: () => Promise<LazyPageModule>
  requiredRoles?: string[]
  navGroup?: string
}

export const ROUTE_DEFINITIONS: RouteDefinition[] = [
  {
    key: 'dashboard',
    path: '/dashboard',
    label: '仪表盘',
    icon: LayoutDashboard,
    lazy: () => import('@/routes/dashboard/index'),
    navGroup: '概览',
  },
  {
    key: 'suppliers',
    path: '/suppliers',
    label: '供应商管理',
    icon: Box,
    lazy: () => import('@/routes/suppliers/index'),
    navGroup: '业务管理',
  },
  {
    key: 'supplierDetail',
    path: '/suppliers/:id',
    label: '供应商详情',
    icon: Box,
    lazy: () => import('@/routes/suppliers/detail'),
  },
  {
    key: 'models',
    path: '/models',
    label: '模型管理',
    icon: Box,
    lazy: () => import('@/routes/models/index'),
    navGroup: '业务管理',
  },
  {
    key: 'contracts',
    path: '/contracts',
    label: '合同管理',
    icon: FileText,
    lazy: () => import('@/routes/contracts/index'),
    navGroup: '业务管理',
  },
  {
    key: 'orders',
    path: '/orders',
    label: '采购订单',
    icon: ShoppingCart,
    lazy: () => import('@/routes/orders/index'),
    navGroup: '业务管理',
  },
  {
    key: 'evaluations',
    path: '/evaluations',
    label: '绩效评估',
    icon: Scale,
    lazy: () => import('@/routes/evaluations/index'),
    navGroup: '业务管理',
  },
  {
    key: 'newapi',
    path: '/newapi',
    label: 'NewAPI',
    icon: Globe,
    lazy: () => import('@/routes/newapi/index'),
    navGroup: '系统设置',
    requiredRoles: ['admin'],
  },
  {
    key: 'users',
    path: '/system/users',
    label: '用户管理',
    icon: Users,
    lazy: () => import('@/routes/system/users'),
    navGroup: '系统设置',
    requiredRoles: ['admin'],
  },
  {
    key: 'weights',
    path: '/system/weights',
    label: '评估权重',
    icon: Settings,
    lazy: () => import('@/routes/system/weights'),
    navGroup: '系统设置',
    requiredRoles: ['admin'],
  },
]

export const NAV_ROUTES = ROUTE_DEFINITIONS.filter((r) => r.navGroup)

export interface NavGroupEntry {
  group: string
  items: RouteDefinition[]
}

export const NAV_GROUP_LAYOUT: NavGroupEntry[] = (() => {
  const groups: NavGroupEntry[] = []
  for (const def of NAV_ROUTES) {
    let group = groups.find((g) => g.group === def.navGroup)
    if (!group) {
      group = { group: def.navGroup!, items: [] }
      groups.push(group)
    }
    group.items.push(def)
  }
  return groups
})()
