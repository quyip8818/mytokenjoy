import { useCallback, useEffect, useState } from 'react'
import { GitBranch } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { routingApi } from '@/api/models'
import type { RoutingRule } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { EmptyState } from '@/components/ui/empty-state'

export default function ModelRoutingPage() {
  const { open } = useWorkflow()
  const [rules, setRules] = useState<RoutingRule[]>([])

  const load = useCallback(async () => {
    const r = await routingApi.getRules()
    setRules(r)
  }, [])

  useEffect(() => {
    void routingApi.getRules().then(setRules)
  }, [])

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
    <div className="space-y-6">
      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          {rules.length === 0 ? (
            <EmptyState
              icon={GitBranch}
              title="暂无路由规则"
              description="组织节点将继承父级的模型白名单配置"
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    组织节点
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    模型范围
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    来源
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground w-[100px]">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rules.map((rule) => (
                  <TableRow key={rule.id} className="border-border/40 hover:bg-indigo-50/30">
                    <TableCell className="font-medium">{rule.nodeName}</TableCell>
                    <TableCell>
                      <span className="text-sm">
                        已选 {rule.allowedModels.length} / 父级 {getParentCount(rule)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={`border-0 ${rule.inherited ? 'bg-slate-100 text-slate-600' : 'bg-indigo-50 text-indigo-700'}`}
                      >
                        {rule.inherited ? '继承' : '自定义'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => open('whitelist-config', { rule, onSuccess: load })}
                      >
                        配置
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
