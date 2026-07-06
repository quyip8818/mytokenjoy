import { useLocation } from 'react-router'

const routeTitles: Record<string, string> = {
  '/dashboard/cost': '成本看板',
  '/dashboard/usage': '用量分析',
  '/org/structure': '组织架构',
  '/org/data-source': '数据源',
  '/org/roles': '角色管理',
  '/budget': '预算管理',
  '/budget/alerts': '预警规则',
  '/keys/provider': '供应商 Key',
  '/keys/platform': '平台凭证',
  '/keys/approval': '审批管理',
  '/models/list': '模型列表',
  '/models/routing': '路由规则',
  '/audit/operations': '操作日志',
  '/audit/calls': '调用日志',
}

export function Header() {
  const location = useLocation()
  const title = routeTitles[location.pathname] || '控制台'

  return (
    <header className="h-14 border-b border-border bg-card flex items-center justify-between px-8">
      <h1 className="text-sm font-medium text-foreground">{title}</h1>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 transition-colors hover:bg-muted">
          <div className="h-6 w-6 rounded-md bg-primary flex items-center justify-center text-[10px] font-medium text-primary-foreground">
            管
          </div>
          <span className="text-sm text-foreground">管理员</span>
        </div>
      </div>
    </header>
  )
}
