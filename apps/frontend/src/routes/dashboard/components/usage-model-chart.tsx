import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import type { ModelUsage } from '@/api/types'

export interface UsageModelChartProps {
  modelUsage: readonly ModelUsage[]
}

export function UsageModelChart({ modelUsage }: UsageModelChartProps) {
  return (
    <ResponsiveContainer width="100%" height={320}>
      <BarChart data={[...modelUsage]} layout="vertical">
        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
        <XAxis type="number" fontSize={11} stroke="#94a3b8" />
        <YAxis type="category" dataKey="modelName" width={130} fontSize={12} stroke="#94a3b8" />
        <Tooltip
          formatter={(value, name) => [
            name === '花费 (¥)'
              ? `¥${Number(value).toLocaleString()}`
              : Number(value).toLocaleString(),
            name,
          ]}
          contentStyle={{
            borderRadius: '8px',
            border: '1px solid #e2e8f0',
            boxShadow: '0 4px 12px rgba(37,99,235,0.08)',
          }}
        />
        <Legend wrapperStyle={{ fontSize: '12px' }} />
        <Bar dataKey="cost" name="花费 (¥)" fill="#2563eb" radius={[0, 4, 4, 0]} />
        <Bar dataKey="requests" name="请求数" fill="#3b82f6" radius={[0, 4, 4, 0]} />
      </BarChart>
    </ResponsiveContainer>
  )
}
