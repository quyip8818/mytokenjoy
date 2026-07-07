import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { DataSection } from '@/components/layout/data-section'
import type { CostGranularity, DailyCost } from '@/api/types'

interface CostTrendChartProps {
  dailyCosts: DailyCost[]
  loading: boolean
  granularity: CostGranularity
}

export function CostTrendChart({ dailyCosts, loading, granularity }: CostTrendChartProps) {
  return (
    <DataSection title="花费趋势" loading={loading} skeletonColumns={1} className="col-span-2">
      <ResponsiveContainer width="100%" height={280}>
        <LineChart data={dailyCosts}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
          <XAxis
            dataKey="date"
            tickFormatter={(v) => (granularity === 'month' ? v : v.slice(5))}
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
  )
}
