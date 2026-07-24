import { Activity, BarChart3, Timer } from 'lucide-react'
import type { ModelRank, NamedValue, TimeSeriesPoint } from '@/api/types/mydashboard'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { MyChartSection } from '@/features/mydashboard'
import { formatMoney } from '@/lib/quota-display'

const CALL_DISTRIBUTION_COLORS = ['#4f46e5', '#10b981', '#f59e0b', '#06b6d4'] as const

interface MyConsumptionChartsProps {
  loading: boolean
  consumptionDistribution: TimeSeriesPoint[]
  consumptionTrend: TimeSeriesPoint[]
  callDistribution: NamedValue[]
  callRanking: ModelRank[]
  distributionTotal: number
  trendTotal: number
  callTotal: number
}

export function MyConsumptionCharts({
  loading,
  consumptionDistribution,
  consumptionTrend,
  callDistribution,
  callRanking,
  distributionTotal,
  trendTotal,
  callTotal,
}: MyConsumptionChartsProps) {
  return (
    <>
      <div className="grid grid-cols-2 gap-4">
        <MyChartSection title="消耗分布" icon={BarChart3}>
          <div>
            <h4 className="text-sm font-medium">模型消耗分布</h4>
            <p className="mb-4 text-xs text-muted-foreground">
              总计：{loading ? '—' : formatMoney(distributionTotal)}
            </p>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={consumptionDistribution}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                <XAxis
                  dataKey="time"
                  fontSize={11}
                  stroke="var(--color-muted-foreground)"
                  tickLine={false}
                />
                <YAxis
                  fontSize={11}
                  stroke="var(--color-muted-foreground)"
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip
                  contentStyle={{
                    borderRadius: '8px',
                    border: '1px solid var(--color-border)',
                    boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#f59e0b"
                  fill="#fef3c7"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </MyChartSection>

        <MyChartSection title="消耗趋势" icon={Activity}>
          <div>
            <h4 className="text-sm font-medium">模型消耗趋势</h4>
            <p className="mb-4 text-xs text-muted-foreground">
              总计：{loading ? '—' : formatMoney(trendTotal)}
            </p>
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={consumptionTrend}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                <XAxis
                  dataKey="time"
                  fontSize={11}
                  stroke="var(--color-muted-foreground)"
                  tickLine={false}
                />
                <YAxis
                  fontSize={11}
                  stroke="var(--color-muted-foreground)"
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip
                  contentStyle={{
                    borderRadius: '8px',
                    border: '1px solid var(--color-border)',
                    boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#f59e0b"
                  strokeWidth={2}
                  dot={{ r: 3, fill: '#f59e0b' }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </MyChartSection>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <MyChartSection title="调用次数分布" icon={Timer}>
          <div>
            <h4 className="text-sm font-medium">模型调用次数占比</h4>
            <p className="mb-4 text-xs text-muted-foreground">
              总计：{loading ? '—' : String(callTotal)}
            </p>
            {callDistribution.length === 0 ? (
              <div className="flex h-[220px] items-center justify-center">
                <p className="text-sm text-muted-foreground">无数据</p>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={220}>
                <PieChart>
                  <Pie
                    data={callDistribution.map((item, index) => ({
                      ...item,
                      fill: CALL_DISTRIBUTION_COLORS[index % CALL_DISTRIBUTION_COLORS.length],
                    }))}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    label
                  />
                  <Legend wrapperStyle={{ fontSize: '12px' }} />
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </MyChartSection>

        <MyChartSection title="调用次数排行" icon={BarChart3}>
          <div>
            <h4 className="text-sm font-medium">模型调用次数排行</h4>
            <p className="mb-4 text-xs text-muted-foreground">
              总计：{loading ? '—' : String(callTotal)}
            </p>
            {callRanking.length === 0 ? (
              <div className="flex h-[220px] items-center justify-center">
                <p className="text-sm text-muted-foreground">无数据</p>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={callRanking} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                  <XAxis type="number" fontSize={11} stroke="var(--color-muted-foreground)" />
                  <YAxis
                    type="category"
                    dataKey="model"
                    fontSize={11}
                    stroke="var(--color-muted-foreground)"
                    width={100}
                  />
                  <Tooltip />
                  <Bar dataKey="count" fill="#4f46e5" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </MyChartSection>
      </div>
    </>
  )
}
