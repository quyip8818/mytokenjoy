import { GitBranch } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useModelRoutingPage } from '@/routes/models/hooks/use-model-routing-page'

export default function ModelRoutingPage() {
  const { rules, loading, error, refresh, getParentCount, openWhitelistConfig } =
    useModelRoutingPage()

  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={4}
        empty={listEmpty(loading, rules, {
          icon: GitBranch,
          title: '暂无路由规则',
          description: '组织节点将继承父级的模型白名单配置',
        })}
      >
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>组织节点</TableHead>
              <TableHead>模型范围</TableHead>
              <TableHead>来源</TableHead>
              <TableHead className="w-[100px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rules.map((rule) => (
              <TableRow key={rule.id}>
                <TableCell className="font-medium">{rule.nodeName}</TableCell>
                <TableCell>
                  <span className="text-sm">
                    已选 {rule.allowedModels.length} / 父级 {getParentCount(rule)}
                  </span>
                </TableCell>
                <TableCell>
                  <StatusBadge variant={rule.inherited ? 'neutral' : 'info'}>
                    {rule.inherited ? '继承' : '自定义'}
                  </StatusBadge>
                </TableCell>
                <TableCell>
                  <PermissionGate write permission={PERMISSION.MODEL_WHITELIST}>
                    <Button variant="ghost" size="sm" onClick={() => openWhitelistConfig(rule)}>
                      配置
                    </Button>
                  </PermissionGate>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
