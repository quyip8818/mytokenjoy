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
import { routingApi } from '@/api/models'
import type { RoutingRule } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { listEmpty } from '@/lib/list-empty'

export default function ModelRoutingPage() {
  const { data: rules = [], loading, refresh } = useAsyncResource(() => routingApi.getRules(), [])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const getParentCount = (rule: RoutingRule) => {
    const parentMap: Record<string, string> = {
      'dept-2': 'dept-1',
      'dept-3': 'dept-2',
      'dept-4': 'dept-2',
      'dept-5': 'dept-2',
      'dept-6': 'dept-1',
      'dept-7': 'dept-1',
      'dept-8': 'dept-1',
    }
    const parentId = parentMap[rule.nodeId]
    const parent = parentId ? rules.find((r) => r.nodeId === parentId) : undefined
    return parent?.allowedModels.length ?? rule.allowedModels.length
  }

  return (
    <PageShell>
      <DataSection
        loading={loading}
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
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => openWithRefresh('whitelist-config', { rule })}
                  >
                    配置
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
