import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import type { OperationDailyCount } from '@/api/types'

interface OperationsTimelineChartProps {
  data: OperationDailyCount[]
  loading: boolean
  onDayClick?: (date: string) => void
}

export function OperationsTimelineChart({ data, loading, onDayClick }: OperationsTimelineChartProps) {
  if (loading) {
    return (
      <div className="flex h-[120px] items-center justify-center rounded-lg border bg-card">
        <div className="h-16 w-full animate-pulse rounded bg-muted mx-6" />
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="flex h-[120px] items-center justify-center rounded-lg border bg-card">
        <p className="text-sm text-muted-foreground">暂无操作记录</p>
      </div>
    )
  }

  const formatted = data.map((d) => ({
    ...d,
    label: d.date.slice(5), // "MM-DD"
  }))

  return (
    <div className="rounded-lg border bg-card p-4">
      <h4 className="mb-2 text-xs font-medium text-muted-foreground">操作时间线</h4>
      <ResponsiveContainer width="100%" height={90}>
        <BarChart
          data={formatted}
          onClick={(_state, event) => {
            if (!onDayClick || !event) return
            // Use activeLabel from the chart event
            const target = event.target as HTMLElement | null
            const label = target?.closest?.('[data-date]')?.getAttribute('data-date')
            if (label) onDayClick(label)
          }}
          style={{ cursor: onDayClick ? 'pointer' : 'default' }}
        >
          <XAxis dataKey="label" tick={{ fontSize: 11 }} axisLine={false} tickLine={false} />
          <YAxis hide allowDecimals={false} />
          <Tooltip
            formatter={(value) => [`${value} 次操作`, '操作数']}
            labelFormatter={(label) => `日期：${label}`}
            contentStyle={{ fontSize: 12 }}
          />
          <Bar dataKey="count" fill="#4f46e5" radius={[2, 2, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
