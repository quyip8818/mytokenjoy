import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { Key, MoreHorizontal } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { providerKeyApi } from '@/api/keys'
import type { ProviderKey } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { EmptyState } from '@/components/ui/empty-state'
import { KeyStatusBadge, ProviderBadge } from '@/lib/label-badges'
import { useRowHighlight } from '@/lib/use-row-highlight'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export default function ProviderKeysPage() {
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const [keys, setKeys] = useState<ProviderKey[]>([])

  const load = useCallback(() => {
    providerKeyApi.list().then(setKeys)
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const handleToggle = async (key: ProviderKey) => {
    const enabled = key.status !== 'active'
    await providerKeyApi.toggle(key.id, enabled)
    toast.success(enabled ? 'Key 已启用' : 'Key 已禁用')
    flashRow(key.id)
    load()
  }

  const handleDelete = async (id: string) => {
    await providerKeyApi.delete(id)
    toast.success('Key 已删除')
    load()
  }

  const getBalanceClass = (balance: number | null) => {
    if (balance === null) return 'text-muted-foreground'
    if (balance > 1000) return 'text-emerald-600 font-medium'
    if (balance < 500) return 'text-amber-600 font-medium'
    return ''
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <Button
          size="sm"
          onClick={() => open('provider-key-form', { onSuccess: load })}
          className="bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white shadow-button hover:-translate-y-0.5 transition-all duration-150"
        >
          添加 Key
        </Button>
      </div>

      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          {keys.length === 0 ? (
            <EmptyState
              icon={Key}
              title="暂无供应商 Key"
              description="添加供应商 Key 后可用于模型路由"
              actionLabel="添加 Key"
              onAction={() => open('provider-key-form', { onSuccess: load })}
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    名称
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    供应商
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    Key 前缀
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    状态
                  </TableHead>
                  <TableHead className="text-right text-xs font-semibold text-muted-foreground">
                    余额
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    最后使用
                  </TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((key) => (
                  <TableRow key={key.id} className={rowClass(key.id)}>
                    <TableCell className="font-medium">{key.name}</TableCell>
                    <TableCell>
                      <ProviderBadge provider={key.provider} />
                    </TableCell>
                    <TableCell>
                      <span className="rounded bg-indigo-50 px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
                        {key.keyPrefix}
                      </span>
                    </TableCell>
                    <TableCell>
                      <KeyStatusBadge status={key.status} />
                    </TableCell>
                    <TableCell className={`text-right ${getBalanceClass(key.balance)}`}>
                      {key.balance !== null ? `$${key.balance.toFixed(2)}` : '-'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {key.lastUsed ?? '-'}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          }
                        />
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleToggle(key)}>
                            {key.status === 'active' ? '禁用' : '启用'}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-red-600"
                            onClick={() => handleDelete(key.id)}
                          >
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
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
