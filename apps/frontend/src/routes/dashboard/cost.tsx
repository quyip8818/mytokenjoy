import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts'
import { dashboardApi } from '@/api/dashboard'
import type { CostSummary, DailyCost, DepartmentCost, TopConsumer } from '@/api/types'
import { TrendingUp, TrendingDown, Coins, Hash, Zap, DollarSign } from 'lucide-react'

const COLORS = ['#4f46e5', '#7c3aed', '#10b981', '#f59e0b', '#06b6d4']

export default function CostDashboardPage() {
  const [summary, setSummary] = useState<CostSummary | null>(null)
  const [dailyCosts, setDailyCosts] = useState<DailyCost[]>([])
  const [deptCosts, setDeptCosts] = useState<DepartmentCost[]>([])
  const [topConsumers, setTopConsumers] = useState<TopConsumer[]>([])

  useEffect(() => {
    Promise.all([
      dashboardApi.getCostSummary(),
      dashboardApi.getDailyCosts(),
      dashboardApi.getDepartmentCosts(),
      dashboardApi.getTopConsumers(),
    ]).then(([s, d, dc, tc]) => {
      setSummary(s)
      setDailyCosts(d)
      setDeptCosts(dc)
      setTopConsumers(tc)
    })
  }, [])

  const stats = [
    { label: '本月总花费', value: summary ? `¥${summary.totalCost.toLocaleString()}` : '-', icon: Coins, accent: 'bg-primary' },
    { label: '环比变化', value: summary ? `${summary.monthOverMonth > 0 ? '+' : ''}${summary.monthOverMonth}%` : '-', icon: summary && summary.monthOverMonth > 0 ? TrendingUp : TrendingDown, accent: summary && summary.monthOverMonth > 0 ? 'bg-red-400' : 'bg-emerald-400' },
    { label: '总 Token 数', value: summary ? `${(summary.totalTokens / 1000000).toFixed(1)}M` : '-', icon: Hash, accent: 'bg-violet-500' },
    { label: '总请求数', value: summary?.totalRequests.toLocaleString() ?? '-', icon: Zap, accent: 'bg-amber-400' },
    { label: '平均请求成本', value: summary ? `¥${summary.avgCostPerRequest.toFixed(2)}` : '-', icon: DollarSign, accent: 'bg-cyan-400' },
  ]

  return (
    <div className="space-y-8">
      {/* Summary stat cards */}
      <div className="grid grid-cols-5 gap-5">
        {stats.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.label} className="shadow-xs border-border">
              <CardContent className="pt-5 pb-4 px-5">
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium text-muted-foreground">{stat.label}</span>
                  <div className={`h-8 w-8 rounded-lg ${stat.accent} flex items-center justify-center`}>
                    <Icon className="h-4 w-4 text-white" />
                  </div>
                </div>
                <div className="text-2xl font-bold tracking-tight">{stat.value}</div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-3 gap-6">
        <Card className="col-span-2 shadow-xs border-border">
          <CardContent className="pt-5 pb-4">
            <h3 className="text-sm font-semibold text-foreground/80 mb-4">每日花费趋势</h3>
            <ResponsiveContainer width="100%" height={280}>
              <LineChart data={dailyCosts}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} fontSize={11} stroke="#94a3b8" />
                <YAxis fontSize={11} stroke="#94a3b8" />
                <Tooltip
                  formatter={(value) => [`¥${Number(value).toFixed(2)}`, '花费']}
                  labelFormatter={(l) => `日期: ${l}`}
                  contentStyle={{ borderRadius: '8px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(79,70,229,0.08)' }}
                />
                <Line type="monotone" dataKey="cost" stroke="#4f46e5" strokeWidth={2.5} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="shadow-xs border-border">
          <CardContent className="pt-5 pb-4">
            <h3 className="text-sm font-semibold text-foreground/80 mb-4">部门成本占比</h3>
            <ResponsiveContainer width="100%" height={280}>
              <PieChart>
                <Pie data={deptCosts} dataKey="cost" nameKey="departmentName" cx="50%" cy="50%" outerRadius={85} label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`} labelLine={false} fontSize={10}>
                  {deptCosts.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Legend wrapperStyle={{ fontSize: '12px' }} />
                <Tooltip formatter={(value) => [`¥${Number(value).toLocaleString()}`, '花费']} contentStyle={{ borderRadius: '8px', border: '1px solid #e2e8f0' }} />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Top consumers */}
      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">消费排行 Top 5</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">排名</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">成员</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">花费 (¥)</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">Token 数</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">请求数</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {topConsumers.map((c, i) => (
                <TableRow key={c.memberId} className="border-border-subtle hover:bg-muted/50">
                  <TableCell>
                    <div className={`h-6 w-6 rounded-full flex items-center justify-center text-[11px] font-bold text-white ${i < 3 ? 'bg-primary' : 'bg-slate-300'}`}>
                      {i + 1}
                    </div>
                  </TableCell>
                  <TableCell className="font-medium">{c.memberName}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">{c.department}</TableCell>
                  <TableCell className="text-right font-semibold">{c.cost.toLocaleString()}</TableCell>
                  <TableCell className="text-right text-muted-foreground">{(c.tokens / 1000000).toFixed(1)}M</TableCell>
                  <TableCell className="text-right text-muted-foreground">{c.requests.toLocaleString()}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
