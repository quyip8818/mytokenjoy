import {
  Wallet, Zap, DollarSign, Activity,
  Send, BarChart3, Coins, Timer,
  Gauge, Clock,
} from 'lucide-react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { ROUTES } from '@/config/routes'
import { PERMISSION } from '@/lib/permissions'
import { usePermissions } from '@/hooks/use-permissions'
import {
  LineChart, Line, AreaChart, Area,
  BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from 'recharts'
import { useMemberDashboardPage } from '@/routes/member/hooks/use-member-dashboard-page'

function StatGroup({
  title,
  icon: Icon,
  items,
  action,
}: {
  title: string
  icon: React.ComponentType<{ className?: string }>
  items: { label: string; value: string; icon: React.ComponentType<{ className?: string }>; action?: React.ReactNode }[]
  action?: React.ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-5 shadow-xs">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Icon className="size-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">{title}</h3>
        </div>
        {action}
      </div>
      <div className="space-y-3">
        {items.map((item) => (
          <div key={item.label} className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-md bg-muted flex items-center justify-center shrink-0">
              <item.icon className="size-4 text-muted-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-xs text-muted-foreground">{item.label}</p>
              <p className="text-base font-semibold tabular-nums">{item.value}</p>
            </div>
            {item.action}
          </div>
        ))}
      </div>
    </div>
  )
}

function ChartSection({
  title,
  icon: Icon,
  children,
}: {
  title: string
  icon: React.ComponentType<{ className?: string }>
  children: React.ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card shadow-xs">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <Icon className="size-4 text-muted-foreground" />
        <h3 className="text-sm font-semibold">{title}</h3>
      </div>
      <div className="p-5">{children}</div>
    </div>
  )
}

export default function MemberDashboardPage() {
  const navigate = useNavigate()
  const { has } = usePermissions()
  const {
    loading,
    accountData,
    usageStats,
    resourceConsumption,
    performance,
    consumptionTrend,
    consumptionDistribution,
    callDistribution,
    callRanking,
    distributionTotal,
    trendTotal,
    callTotal,
  } = useMemberDashboardPage()

  const handleRecharge = () => {
    if (has([PERMISSION.BILLING_READ, PERMISSION.BILLING_RECHARGE])) {
      navigate(ROUTES.wallet)
      return
    }
    toast.message('请联系管理员进行充值')
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-4 gap-4">
        <StatGroup
          title="账户数据"
          icon={Wallet}
          items={[
            {
              label: '当前余额',
              value: loading ? '—' : `¥${accountData.balance.toFixed(2)}`,
              icon: Coins,
              action: (
                <Button
                  variant="outline"
                  size="sm"
                  className="h-6 text-xs px-2"
                  onClick={handleRecharge}
                >
                  充值
                </Button>
              ),
            },
            {
              label: '历史消耗',
              value: loading ? '—' : `¥${accountData.totalSpent.toFixed(2)}`,
              icon: DollarSign,
            },
          ]}
        />
        <StatGroup
          title="使用统计"
          icon={Zap}
          items={[
            {
              label: '请求次数',
              value: loading ? '—' : String(usageStats.requestCount),
              icon: Send,
            },
            {
              label: '统计次数',
              value: loading ? '—' : String(usageStats.totalCount),
              icon: BarChart3,
            },
          ]}
        />
        <StatGroup
          title="资源消耗"
          icon={DollarSign}
          items={[
            {
              label: '统计额度',
              value: loading ? '—' : `¥${resourceConsumption.totalCost.toFixed(2)}`,
              icon: Coins,
            },
            {
              label: '统计 Tokens',
              value: loading ? '—' : String(resourceConsumption.totalTokens),
              icon: Activity,
            },
          ]}
        />
        <StatGroup
          title="性能指标"
          icon={Activity}
          items={[
            {
              label: '平均 RPM',
              value: loading ? '—' : performance.avgRPM.toFixed(3),
              icon: Gauge,
            },
            {
              label: '平均 TPM',
              value: loading ? '—' : String(performance.avgTPM),
              icon: Clock,
            },
          ]}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <ChartSection title="消耗分布" icon={BarChart3}>
          <div>
            <h4 className="text-sm font-medium">模型消耗分布</h4>
            <p className="text-xs text-muted-foreground mb-4">
              总计：{loading ? '—' : `¥${distributionTotal.toFixed(2)}`}
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
        </ChartSection>

        <ChartSection title="消耗趋势" icon={Activity}>
          <div>
            <h4 className="text-sm font-medium">模型消耗趋势</h4>
            <p className="text-xs text-muted-foreground mb-4">
              总计：{loading ? '—' : trendTotal.toFixed(2)}
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
        </ChartSection>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <ChartSection title="调用次数分布" icon={Timer}>
          <div>
            <h4 className="text-sm font-medium">模型调用次数占比</h4>
            <p className="text-xs text-muted-foreground mb-4">
              总计：{loading ? '—' : String(callTotal)}
            </p>
            {callDistribution.length === 0 ? (
              <div className="flex items-center justify-center h-[220px]">
                <p className="text-sm text-muted-foreground">无数据</p>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={220}>
                <PieChart>
                  <Pie
                    data={callDistribution}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    label
                  >
                    {callDistribution.map((_, i) => (
                      <Cell key={i} fill={['#4f46e5', '#10b981', '#f59e0b', '#06b6d4'][i % 4]} />
                    ))}
                  </Pie>
                  <Legend wrapperStyle={{ fontSize: '12px' }} />
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </ChartSection>

        <ChartSection title="调用次数排行" icon={BarChart3}>
          <div>
            <h4 className="text-sm font-medium">模型调用次数排行</h4>
            <p className="text-xs text-muted-foreground mb-4">
              总计：{loading ? '—' : String(callTotal)}
            </p>
            {callRanking.length === 0 ? (
              <div className="flex items-center justify-center h-[220px]">
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
        </ChartSection>
      </div>
    </div>
  )
}
