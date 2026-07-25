import { useInjectedQuery, queryKeys } from '@/features/query'
import { ORDER_STATUS } from '@/config/enums'
import { Badge } from '@/components/ui'

export function DashboardPage() {
  const { data: cards } = useInjectedQuery({
    queryKey: queryKeys.dashboard.cards(),
    queryFn: (a) => a.dashboardApi.cards(),
  })
  const { data: expiring } = useInjectedQuery({
    queryKey: queryKeys.dashboard.expiring(),
    queryFn: (a) => a.dashboardApi.expiring(),
  })
  const { data: recentOrders } = useInjectedQuery({
    queryKey: queryKeys.dashboard.recentOrders(),
    queryFn: (a) => a.dashboardApi.recentOrders(),
  })
  const { data: charts } = useInjectedQuery({
    queryKey: queryKeys.dashboard.charts(),
    queryFn: (a) => a.dashboardApi.charts(),
  })

  const statCards = [
    { label: '供应商总数', value: cards?.supplierTotal ?? 0, color: 'bg-blue-500' },
    { label: '合作中供应商', value: cards?.activeSuppliers ?? 0, color: 'bg-green-500' },
    { label: '模型总数', value: cards?.modelTotal ?? 0, color: 'bg-purple-500' },
    { label: '生效合同', value: cards?.activeContracts ?? 0, color: 'bg-yellow-500' },
  ]

  return (
    <div className="space-y-6">
      {/* 统计卡片 */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((card) => (
          <div key={card.label} className="rounded-lg border bg-white p-5">
            <div className="flex items-center gap-3">
              <div
                className={`h-10 w-10 rounded-lg ${card.color} flex items-center justify-center`}
              >
                <span className="text-lg font-bold text-white">{card.value}</span>
              </div>
              <div>
                <div className="text-2xl font-bold">{card.value}</div>
                <div className="text-sm text-muted-foreground">{card.label}</div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* 即将到期合同 */}
        <div className="rounded-lg border bg-white p-5">
          <h2 className="mb-3 text-sm font-semibold">30 天内到期合同</h2>
          {!expiring || expiring.length === 0 ? (
            <div className="py-6 text-center text-sm text-muted-foreground">暂无即将到期的合同</div>
          ) : (
            <div className="space-y-2">
              {expiring.map((c) => (
                <div
                  key={c.id}
                  className="flex items-center justify-between rounded border px-3 py-2 text-sm"
                >
                  <div>
                    <span className="font-medium">{c.title}</span>
                    <span className="ml-2 text-muted-foreground">{c.supplierName}</span>
                  </div>
                  <span className="text-xs font-medium text-yellow-600">{c.endDate}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* 最近订单 */}
        <div className="rounded-lg border bg-white p-5">
          <h2 className="mb-3 text-sm font-semibold">最近订单</h2>
          {!recentOrders || recentOrders.length === 0 ? (
            <div className="py-6 text-center text-sm text-muted-foreground">暂无订单</div>
          ) : (
            <div className="space-y-2">
              {recentOrders.map((o) => (
                <div
                  key={o.id}
                  className="flex items-center justify-between rounded border px-3 py-2 text-sm"
                >
                  <div>
                    <span className="font-medium">{o.orderNo}</span>
                    <span className="ml-2 text-muted-foreground">{o.supplierName}</span>
                  </div>
                  <Badge
                    variant={
                      o.status === 'completed'
                        ? 'success'
                        : o.status === 'pending'
                          ? 'warning'
                          : 'default'
                    }
                  >
                    {ORDER_STATUS[o.status]?.label ?? o.status}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 图表区 */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-lg border bg-white p-5">
          <h2 className="mb-3 text-sm font-semibold">评估等级分布</h2>
          {charts?.gradeDistribution && charts.gradeDistribution.length > 0 ? (
            <div className="flex h-32 items-end gap-3">
              {charts.gradeDistribution.map((item) => {
                const max = Math.max(...charts.gradeDistribution.map((d) => d.count), 1)
                const h = (item.count / max) * 100
                const colors: Record<string, string> = {
                  A: 'bg-green-500',
                  B: 'bg-blue-500',
                  C: 'bg-yellow-500',
                  D: 'bg-red-500',
                }
                return (
                  <div key={item.label} className="flex flex-1 flex-col items-center gap-1">
                    <span className="text-xs font-medium">{item.count}</span>
                    <div
                      className={`w-full rounded-t ${colors[item.label] ?? 'bg-gray-400'}`}
                      style={{ height: `${h}%`, minHeight: 4 }}
                    />
                    <span className="text-xs text-muted-foreground">{item.label}</span>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="py-8 text-center text-sm text-muted-foreground">暂无数据</div>
          )}
        </div>

        <div className="rounded-lg border bg-white p-5">
          <h2 className="mb-3 text-sm font-semibold">各供应商模型数量</h2>
          {charts?.modelCountBySupplier && charts.modelCountBySupplier.length > 0 ? (
            <div className="space-y-2">
              {charts.modelCountBySupplier.slice(0, 8).map((item) => {
                const max = Math.max(...charts.modelCountBySupplier.map((d) => d.count), 1)
                return (
                  <div key={item.label} className="flex items-center gap-2 text-sm">
                    <span className="w-20 truncate text-muted-foreground">{item.label}</span>
                    <div className="h-4 flex-1 overflow-hidden rounded bg-muted">
                      <div
                        className="h-full rounded bg-blue-500"
                        style={{ width: `${(item.count / max) * 100}%` }}
                      />
                    </div>
                    <span className="w-6 text-right text-xs font-medium">{item.count}</span>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="py-8 text-center text-sm text-muted-foreground">暂无数据</div>
          )}
        </div>
      </div>
    </div>
  )
}
