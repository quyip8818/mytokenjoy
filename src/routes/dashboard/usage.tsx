import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { BudgetProgressCell } from '@/components/ui/budget-progress-cell'
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
import { dashboardApi } from '@/api/dashboard'
import { useAsyncResource } from '@/hooks/use-async-resource'

export default function UsageDashboardPage() {
  const { data, loading } = useAsyncResource(async () => {
    const [teamUsage, modelUsage] = await Promise.all([
      dashboardApi.getTeamUsage(),
      dashboardApi.getModelUsage(),
    ])
    return { teamUsage, modelUsage }
  }, [])

  const teamUsage = data?.teamUsage ?? []
  const modelUsage = data?.modelUsage ?? []

  return (
    <PageShell>
      <DataSection title="团队用量" loading={loading} skeletonColumns={6}>
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>部门</TableHead>
              <TableHead>额度 (¥)</TableHead>
              <TableHead>已消耗 (¥)</TableHead>
              <TableHead className="w-48">消耗进度</TableHead>
              <TableHead className="text-right">成员数</TableHead>
              <TableHead>主力模型</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {teamUsage.map((t) => (
              <TableRow key={t.departmentId}>
                <TableCell className="font-medium">{t.departmentName}</TableCell>
                <TableCell className="text-muted-foreground">{t.quota.toLocaleString()}</TableCell>
                <TableCell className="font-medium">{t.consumed.toLocaleString()}</TableCell>
                <TableCell>
                  <BudgetProgressCell
                    value={t.consumed}
                    total={t.quota}
                    className="gap-2.5"
                    accentLabel
                  />
                </TableCell>
                <TableCell className="text-right text-muted-foreground">{t.memberCount}</TableCell>
                <TableCell>
                  <StatusBadge variant="info">{t.topModel}</StatusBadge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>

      <DataSection title="模型使用分布" loading={loading} skeletonColumns={1}>
        <ResponsiveContainer width="100%" height={320}>
          <BarChart data={modelUsage} layout="vertical">
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
                boxShadow: '0 4px 12px rgba(79,70,229,0.08)',
              }}
            />
            <Legend wrapperStyle={{ fontSize: '12px' }} />
            <Bar dataKey="cost" name="花费 (¥)" fill="#4f46e5" radius={[0, 4, 4, 0]} />
            <Bar dataKey="requests" name="请求数" fill="#7c3aed" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </DataSection>
    </PageShell>
  )
}
