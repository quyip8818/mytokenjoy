import { NavLink } from 'react-router'
import { cn } from '@/lib/utils'
import {
  BarChart3,
  TrendingUp,
  Building2,
  Database,
  Shield,
  Wallet,
  Bell,
  Key,
  CreditCard,
  CheckCircle2,
  Cpu,
  GitBranch,
  FileText,
  Activity,
} from 'lucide-react'

const navItems = [
  {
    group: '数据看板',
    items: [
      { label: '成本看板', path: '/dashboard/cost', icon: BarChart3 },
      { label: '用量分析', path: '/dashboard/usage', icon: TrendingUp },
    ],
  },
  {
    group: '组织管理',
    items: [
      { label: '组织架构', path: '/org/structure', icon: Building2 },
      { label: '数据源', path: '/org/data-source', icon: Database },
      { label: '角色管理', path: '/org/roles', icon: Shield },
    ],
  },
  {
    group: '预算管理',
    items: [
      { label: '预算管理', path: '/budget', icon: Wallet },
      { label: '预警规则', path: '/budget/alerts', icon: Bell },
    ],
  },
  {
    group: 'Key 管理',
    items: [
      { label: '供应商 Key', path: '/keys/provider', icon: Key },
      { label: '平台凭证', path: '/keys/platform', icon: CreditCard },
      { label: '审批管理', path: '/keys/approval', icon: CheckCircle2 },
    ],
  },
  {
    group: '模型路由',
    items: [
      { label: '模型列表', path: '/models/list', icon: Cpu },
      { label: '路由规则', path: '/models/routing', icon: GitBranch },
    ],
  },
  {
    group: '审计日志',
    items: [
      { label: '操作日志', path: '/audit/operations', icon: FileText },
      { label: '调用日志', path: '/audit/calls', icon: Activity },
    ],
  },
  {
    group: '财务管理',
    items: [
      { label: '钱包管理', path: '/wallet', icon: Wallet },
    ],
  },
]

export function Sidebar() {
  return (
    <aside className="w-56 flex flex-col border-r border-border bg-card">
      {/* Logo */}
      <div className="px-5 pt-5 pb-4">
        <h1 className="text-base font-semibold text-primary">TokenJoy</h1>
        <p className="mt-0.5 text-xs text-muted-foreground">LLM API 管理平台</p>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto px-3 pb-4 space-y-5">
        {navItems.map((group) => (
          <div key={group.group}>
            <div className="px-2 mb-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {group.group}
            </div>
            <div className="space-y-0.5">
              {group.items.map((item) => {
                const Icon = item.icon
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    end
                    className={({ isActive }) =>
                      cn(
                        'flex items-center gap-2.5 px-3 py-2 text-sm rounded-md transition-colors duration-100',
                        isActive
                          ? 'bg-muted text-foreground font-medium'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted',
                      )
                    }
                  >
                    <Icon className="h-4 w-4 shrink-0" strokeWidth={1.5} />
                    {item.label}
                  </NavLink>
                )
              })}
            </div>
          </div>
        ))}
      </nav>
    </aside>
  )
}
