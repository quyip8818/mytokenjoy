import { useState, useMemo, useCallback } from 'react'
import { ChevronRight, Users } from 'lucide-react'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
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
import type { CostPeriod, DepartmentCost, DepartmentCostMember } from '@/api/types'
import { COST_PERIOD, COST_PERIOD_LABELS } from '@/lib/dashboard-constants'
import { TrendingUp, TrendingDown, Coins, Hash, Zap, DollarSign, User } from 'lucide-react'

const COLORS = ['#2563eb', '#3b82f6', '#10b981', '#f59e0b', '#06b6d4']

type DrillLevel = 'departments' | 'members'

interface DrillState {
  level: DrillLevel
  parentId: string | null
  parentName: string | null
  deptId: string | null
  deptName: string | null
}

const ROOT_DRILL: DrillState = {
  level: 'departments',
  parentId: null,
  parentName: null,
  deptId: null,
  deptName: null,
}

export default function CostDashboardPage() {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [drill, setDrill] = useState<DrillState>(ROOT_DRILL)

  const { data, loading } = useAsyncResource(async () => {
    const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
      dashboardApi.getCostSummary(period),
      dashboardApi.getDailyCosts(period),
      drill.level === 'members' && drill.deptId
        ? dashboardApi.getDepartmentMemberCosts(drill.deptId, period)
        : dashboardApi.getDepartmentCosts({
            parentId: drill.parentId ?? undefined,
            period,
          }),
      dashboardApi.getTopConsumers({ limit: 5, period }),
    ])
    return { summary, dailyCosts, deptCosts, topConsumers }
  }, [period, drill])

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
    setDrill(ROOT_DRILL)
  }, [])

  const handleDrillDept = useCallback(
    (dept: DepartmentCost) => {
      if (drill.level === 'departments' && dept.hasChildren) {
        setDrill({
          level: 'departments',
          parentId: dept.departmentId,
          parentName: dept.departmentName,
          deptId: null,
          deptName: null,
        })
        return
      }
      if (drill.level === 'departments') {
        setDrill({
          level: 'members',
          parentId: drill.parentId,
          parentName: drill.parentName,
          deptId: dept.departmentId,
          deptName: dept.departmentName,
        })
      }
    },
    [drill],
  )

  const handleDrillBack = useCallback(() => {
    if (drill.level === 'members') {
      setDrill({
        level: 'departments',
        parentId: drill.parentId,
        parentName: drill.parentName,
        deptId: null,
        deptName: null,
      })
      return
    }
    if (drill.parentId) {
      setDrill(ROOT_DRILL)
    }
  }, [drill])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const deptCosts = (data?.deptCosts ?? []) as DepartmentCost[]
  const memberCosts = (data?.deptCosts ?? []) as DepartmentCostMember[]

  const deptCostsWithColors = useMemo(() => {
    if (drill.level === 'members') {
      return memberCosts.map((item, i) => ({
        departmentName: item.memberName,
        cost: item.cost,
        fill: COLORS[i % COLORS.length],
      }))
    }
    return deptCosts.map((item, i) => ({ ...item, fill: COLORS[i % COLORS.length] }))
  }, [deptCosts, memberCosts, drill.level])

  const drillTitle = useMemo(() => {
    if (drill.level === 'members' && drill.deptName) return `${drill.deptName} · 成员明细`
    if (drill.parentName) return `${drill.parentName} · 子部门`
    return '部门花费明细'
  }, [drill])

  const stats = [
    {
      label: '总花费',
      value: summary ? `¥${summary.totalCost.toLocaleString()}` : '-',
      icon: Coins,
      accent: 'from-blue-500 to-sky-500',
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
      label: '人均成本',
      value: summary ? `¥${summary.avgCostPerMember.toLocaleString()}` : '-',
      icon: User,
      accent: 'from-violet-400 to-purple-500',
    },
    {
      label: '总调用次数',
      value: summary?.totalRequests.toLocaleString() ?? '-',
      icon: Zap,
      accent: 'from-amber-400 to-orange-500',
    },
    {
      label: '平均单次成本',
      value: summary ? `¥${summary.avgCostPerRequest.toFixed(2)}` : '-',
      icon: DollarSign,
      accent: 'from-cyan-400 to-blue-500',
    },
    {
      label: '总 Token',
      value: summary ? `${(summary.totalTokens / 1000000).toFixed(1)}M` : '-',
      icon: Hash,
      accent: 'from-blue-500 to-sky-400',
    },
  ]

  const periodActions = (
    <Select value={period} onValueChange={handlePeriodChange}>
      <SelectTrigger className="w-32 border-border/60">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {Object.entries(COST_PERIOD_LABELS).map(([value, label]) => (
          <SelectItem key={value} value={value}>
            {label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )

  return (
    <PageShell
      actions={periodActions}
      stats={
        <div className="grid grid-cols-2 gap-5 lg:grid-cols-6">
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
        <DataSection title="花费趋势" loading={loading} skeletonColumns={1} className="col-span-2">
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
                  boxShadow: '0 4px 12px rgba(37,99,235,0.08)',
                }}
              />
              <Line type="monotone" dataKey="cost" stroke="#2563eb" strokeWidth={2.5} dot={false} />
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

      <DataSection
        title={drillTitle}
        loading={loading}
        skeletonColumns={5}
        headerAction={
          drill.parentId || drill.level === 'members' ? (
            <Button variant="outline" size="sm" onClick={handleDrillBack}>
              返回上级
            </Button>
          ) : undefined
        }
      >
        {drill.level === 'members' ? (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>成员</TableHead>
                <TableHead className="text-right">花费 (¥)</TableHead>
                <TableHead className="text-right">Token 数</TableHead>
                <TableHead className="text-right">请求数</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {memberCosts.map((m) => (
                <TableRow key={m.memberId}>
                  <TableCell className="font-medium">
                    <Users className="mr-2 inline h-4 w-4 text-muted-foreground" />
                    {m.memberName}
                  </TableCell>
                  <TableCell className="text-right font-semibold">
                    {m.cost.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {(m.tokens / 1000000).toFixed(1)}M
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {m.requests.toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-6" />
                <TableHead>部门</TableHead>
                <TableHead className="text-right">花费 (¥)</TableHead>
                <TableHead className="text-right">占比</TableHead>
                <TableHead className="w-24" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {deptCosts.map((dept) => (
                <TableRow key={dept.departmentId}>
                  <TableCell />
                  <TableCell className="font-medium">{dept.departmentName}</TableCell>
                  <TableCell className="text-right font-semibold">
                    {dept.cost.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {dept.percentage}%
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 text-blue-600"
                      onClick={() => handleDrillDept(dept)}
                    >
                      下钻
                      <ChevronRight className="ml-1 h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </DataSection>

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
                    className={`flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white ${i < 3 ? 'bg-gradient-to-br from-blue-500 to-sky-500' : 'bg-slate-300'}`}
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
