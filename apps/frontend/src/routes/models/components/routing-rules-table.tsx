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
