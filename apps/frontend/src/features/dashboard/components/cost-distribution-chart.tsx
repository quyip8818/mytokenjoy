import { PieChart, Pie, Legend, Tooltip, ResponsiveContainer } from 'recharts'
import { DataSection } from '@/components/layout/data-section'

interface CostDistributionChartProps {
  data: Array<{ departmentName: string; cost: number; fill: string }>
  loading: boolean
}

export function CostDistributionChart({ data, loading }: CostDistributionChartProps) {
  return (
    <DataSection title="部门成本占比" loading={loading} skeletonColumns={1}>
      <ResponsiveContainer width="100%" height={280}>
        <PieChart>
          <Pie
            data={data}
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
  )
}
