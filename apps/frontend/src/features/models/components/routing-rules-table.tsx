import type { RoutingRule } from '@/api/types'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/ui/status-badge'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'

interface RoutingRulesTableProps {
  rules: RoutingRule[]
  getParentCount: (rule: RoutingRule) => number
  onConfigure: (rule: RoutingRule) => void
}

export function RoutingRulesTable({ rules, getParentCount, onConfigure }: RoutingRulesTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            组织节点
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            可用模型
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            默认模型
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            降级模型
          </TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">
            来源
          </TableHead>
          <TableHead className="w-[100px] text-xs font-medium uppercase text-muted-foreground">
            操作
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rules.map((rule) => (
          <TableRow key={rule.id} className="even:bg-muted/40">
            <TableCell className="font-medium">{rule.nodeName}</TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {rule.allowedModels.slice(0, 3).map((model) => (
                  <StatusBadge key={model} variant="info" className="text-xs">
                    {model}
                  </StatusBadge>
                ))}
                {rule.allowedModels.length > 3 && (
                  <StatusBadge variant="info" className="text-xs">
                    +{rule.allowedModels.length - 3}
                  </StatusBadge>
                )}
                {rule.allowedModels.length === 0 && (
                  <span className="text-sm text-muted-foreground">
                    父级 {getParentCount(rule)} 个
                  </span>
                )}
              </div>
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {rule.defaultModel ?? '—'}
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {rule.fallbackModel ?? '—'}
            </TableCell>
            <TableCell>
              <StatusBadge variant={rule.inherited ? 'neutral' : 'info'}>
                {rule.inherited ? '继承' : '自定义'}
              </StatusBadge>
            </TableCell>
            <TableCell>
              <PermissionGate write permission={PERMISSION.MODEL_WHITELIST}>
                <Button variant="ghost" size="sm" onClick={() => onConfigure(rule)}>
                  配置
                </Button>
              </PermissionGate>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
