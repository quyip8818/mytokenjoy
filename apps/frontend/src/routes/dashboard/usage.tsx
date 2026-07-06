import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { useUsageDashboardPage } from '@/routes/dashboard/hooks/use-usage-dashboard-page'

export default function UsageDashboardPage() {
  const { teamUsage, modelUsage } = useUsageDashboardPage()

  return (
    <div className="space-y-8">
      {/* Team usage */}
      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">团队用量</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">额度 (¥)</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">已消耗 (¥)</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground w-48">消耗进度</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">成员数</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">主力模型</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {teamUsage.map((t) => {
                const pct = Math.round((t.consumed / t.quota) * 100)
                return (
                  <TableRow key={t.departmentId} className="border-border-subtle hover:bg-muted/50">
                    <TableCell className="font-medium">{t.departmentName}</TableCell>
                    <TableCell className="text-muted-foreground">{t.quota.toLocaleString()}</TableCell>
                    <TableCell className="font-medium">{t.consumed.toLocaleString()}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2.5">
                        <Progress value={pct} className="flex-1 h-2" />
                        <span className={`text-xs font-semibold ${pct >= 90 ? 'text-red-500' : pct >= 70 ? 'text-amber-500' : 'text-primary'}`}>{pct}%</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">{t.memberCount}</TableCell>
                    <TableCell><Badge variant="secondary" className="text-xs font-medium">{t.topModel}</Badge></TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Model usage chart */}
      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">模型使用分布</h3>
          <ResponsiveContainer width="100%" height={320}>
            <BarChart data={modelUsage} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis type="number" fontSize={11} stroke="#94a3b8" />
              <YAxis type="category" dataKey="modelName" width={130} fontSize={12} stroke="#94a3b8" />
              <Tooltip
                formatter={(value, name) => [name === '花费 (¥)' ? `¥${Number(value).toLocaleString()}` : Number(value).toLocaleString(), name]}
                contentStyle={{ borderRadius: '8px', border: '1px solid #e2e8f0', boxShadow: '0 4px 12px rgba(79,70,229,0.08)' }}
              />
              <Legend wrapperStyle={{ fontSize: '12px' }} />
              <Bar dataKey="cost" name="花费 (¥)" fill="#4f46e5" radius={[0, 4, 4, 0]} />
              <Bar dataKey="requests" name="请求数" fill="#7c3aed" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  )
}
