import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { UsageSeriesChartPoint } from '@/lib/dashboard'

interface UsageSeriesChartProps {
  data: UsageSeriesChartPoint[]
}

export function UsageSeriesChart({ data }: UsageSeriesChartProps) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
        <XAxis dataKey="label" fontSize={11} stroke="#94a3b8" />
        <YAxis fontSize={11} stroke="#94a3b8" />
        <Tooltip
          formatter={(value, name) => {
            if (name === 'costCny') return [`¥${Number(value).toFixed(2)}`, '花费']
            return [value, '调用次数']
          }}
          labelFormatter={(label) => `时间: ${label}`}
          contentStyle={{
            borderRadius: '8px',
            border: '1px solid #e2e8f0',
            boxShadow: '0 4px 12px rgba(37,99,235,0.08)',
          }}
        />
        <Line type="monotone" dataKey="costCny" stroke="#2563eb" strokeWidth={2.5} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  )
}
