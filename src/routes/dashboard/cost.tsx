import { useMemo } from 'react'
import { StatCard } from '@/components/ui/stat-card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Legend,
} from 'recharts'
import { dashboardApi } from '@/api/dashboard'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { TrendingUp, TrendingDown, Coins, Hash, Zap, DollarSign } from 'lucide-react'

const COLORS = ['#4f46e5', '#7c3aed', '#10b981', '#f59e0b', '#06b6d4']

export default function CostDashboardPage() {
  const { data, loading } = useAsyncResource(async () => {
    const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
      dashboardApi.getCostSummary(),
      dashboardApi.getDailyCosts(),
      dashboardApi.getDepartmentCosts(),
      dashboardApi.getTopConsumers(),
    ])
    return { summary, dailyCosts, deptCosts, topConsumers }
  }, [])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const deptCostsWithColors = useMemo(() => {
    const items = data?.deptCosts ?? []
    return items.map((item, i) => ({ ...item, fill: COLORS[i % COLORS.length] }))
  }, [data?.deptCosts])

  const stats = [
    {
      label: '本月总花费',
      value: summary ? `¥${summary.totalCost.toLocaleString()}` : '-',
      icon: Coins,
      accent: 'from-indigo-500 to-violet-500',
    },
    {
      label: '环比变化',
      value: summary ? `${summary.monthOverMonth > 0 ? '+' : ''}${summary.monthOverMonth}%` : '-',
      icon: summary && summary.monthOverMonth > 0 ? TrendingUp : TrendingDown,
      accent:
        summary && summary.monthOverMonth > 0
          ? 'from-red-400 to-rose-500'
          : 'from-emerald-400 to-teal-500',
    },
    {
      label: '总 Token 数',
      value: summary ? `${(summary.totalTokens / 1000000).toFixed(1)}M` : '-',
      icon: Hash,
      accent: 'from-violet-500 to-purple-500',
    },
    {
      label: '总请求数',
      value: summary?.totalRequests.toLocaleString() ?? '-',
      icon: Zap,
      accent: 'from-amber-400 to-orange-500',
    },
    {
      label: '平均请求成本',
      value: summary ? `¥${summary.avgCostPerRequest.toFixed(2)}` : '-',
      icon: DollarSign,
      accent: 'from-cyan-400 to-blue-500',
    },
  ]

  return (
    <PageShell
      stats={
        <div className="grid grid-cols-2 gap-5 lg:grid-cols-5">
          {stats.map((stat) => (
            <StatCard
              key={stat.label}
              label={stat.label}
              value={loading ? '-' : stat.value}
              icon={stat.icon}
              iconAccent={stat.accent}
            />
          ))}
        </div>
      }
    >
      <div className="grid grid-cols-3 gap-6">
        <DataSection
          title="每日花费趋势"
          loading={loading}
          skeletonColumns={1}
          className="col-span-2"
        >
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={dailyCosts}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis
                dataKey="date"
                tickFormatter={(v) => v.slice(5)}
                fontSize={11}
                stroke="#94a3b8"
              />
              <YAxis fontSize={11} stroke="#94a3b8" />
              <Tooltip
                formatter={(value) => [`¥${Number(value).toFixed(2)}`, '花费']}
                labelFormatter={(l) => `日期: ${l}`}
                contentStyle={{
                  borderRadius: '8px',
                  border: '1px solid #e2e8f0',
                  boxShadow: '0 4px 12px rgba(79,70,229,0.08)',
                }}
              />
              <Line type="monotone" dataKey="cost" stroke="#4f46e5" strokeWidth={2.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </DataSection>

        <DataSection title="部门成本占比" loading={loading} skeletonColumns={1}>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie
                data={deptCostsWithColors}
                dataKey="cost"
                nameKey="departmentName"
                cx="50%"
                cy="50%"
                outerRadius={85}
                label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                labelLine={false}
                fontSize={10}
              />
              <Legend wrapperStyle={{ fontSize: '12px' }} />
              <Tooltip
                formatter={(value) => [`¥${Number(value).toLocaleString()}`, '花费']}
                contentStyle={{ borderRadius: '8px', border: '1px solid #e2e8f0' }}
              />
            </PieChart>
          </ResponsiveContainer>
        </DataSection>
      </div>

      <DataSection title="消费排行 Top 5" loading={loading} skeletonColumns={6}>
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>排名</TableHead>
              <TableHead>成员</TableHead>
              <TableHead>部门</TableHead>
              <TableHead className="text-right">花费 (¥)</TableHead>
              <TableHead className="text-right">Token 数</TableHead>
              <TableHead className="text-right">请求数</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {topConsumers.map((c, i) => (
              <TableRow key={c.memberId}>
                <TableCell>
                  <div
                    className={`flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white ${i < 3 ? 'bg-gradient-to-br from-indigo-500 to-violet-500' : 'bg-slate-300'}`}
                  >
                    {i + 1}
                  </div>
                </TableCell>
                <TableCell className="font-medium">{c.memberName}</TableCell>
                <TableCell className="text-sm text-muted-foreground">{c.department}</TableCell>
                <TableCell className="text-right font-semibold">
                  {c.cost.toLocaleString()}
                </TableCell>
                <TableCell className="text-right text-muted-foreground">
                  {(c.tokens / 1000000).toFixed(1)}M
                </TableCell>
                <TableCell className="text-right text-muted-foreground">
                  {c.requests.toLocaleString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
