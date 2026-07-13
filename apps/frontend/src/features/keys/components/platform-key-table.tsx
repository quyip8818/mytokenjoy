import type { PlatformKey } from '@/api/types'
import type { PlatformKeyTab } from '@/features/keys'
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
import { BudgetProgressCell } from '@/features/budget'
import { KeyPrefixBadge, KeyStatusBadge } from './status-badges'

interface PlatformKeyTableProps {
  keys: PlatformKey[]
  type: PlatformKeyTab
  rowClass: (id: string) => string
  onRevoke: (id: string) => void
  modelLabel?: (modelId: number) => string
}

export function PlatformKeyTable({
  keys,
  type,
  rowClass,
  onRevoke,
  modelLabel = (id) => `#${id}`,
}: PlatformKeyTableProps) {
  const ownerLabel = type === 'member' ? '成员' : type === 'project_member' ? '成员 / 项目' : '项目'

  return (
    <Table>
      <TableHeader>
        <TableRow className="border-border/50 hover:bg-transparent">
          <TableHead className="text-xs font-medium text-muted-foreground">{ownerLabel}</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">Key 名称</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">所属部门</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">Key 前缀</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">状态</TableHead>
          <TableHead className="w-40 text-xs font-medium text-muted-foreground">额度</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">模型白名单</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">到期时间</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {keys.length === 0 ? (
          <TableRow>
            <TableCell colSpan={9} className="h-32 text-center text-sm text-muted-foreground">
              暂无 Key
            </TableCell>
          </TableRow>
        ) : (
          keys.map((key) => (
            <TableRow
              key={key.id}
              className={`border-border-subtle even:bg-muted/40 hover:bg-muted/50 ${rowClass(key.id)}`}
            >
              <TableCell className="text-sm font-medium">
                {type === 'member'
                  ? (key.memberName ?? '—')
                  : type === 'project_member'
                    ? `${key.memberName ?? '—'} / ${key.projectName ?? '—'}`
                    : (key.projectName ?? '—')}
              </TableCell>
              <TableCell className="text-sm">{key.name}</TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {key.departmentName || '—'}
              </TableCell>
              <TableCell>
                <KeyPrefixBadge prefix={key.keyPrefix} />
              </TableCell>
              <TableCell>
                <KeyStatusBadge status={key.status} />
              </TableCell>
              <TableCell>
                <BudgetProgressCell value={key.consumed} total={key.budget} />
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {key.modelWhitelist.slice(0, 2).map((modelId) => (
                    <StatusBadge key={modelId} variant="info" className="text-xs">
                      {modelLabel(modelId)}
                    </StatusBadge>
                  ))}
                  {key.modelWhitelist.length > 2 && (
                    <StatusBadge variant="info" className="text-xs">
                      +{key.modelWhitelist.length - 2}
                    </StatusBadge>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {key.expiresAt ?? '永不'}
              </TableCell>
              <TableCell>
                {key.status === 'active' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 text-red-600 hover:text-red-700"
                    onClick={() => onRevoke(key.id)}
                  >
                    吊销
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}
