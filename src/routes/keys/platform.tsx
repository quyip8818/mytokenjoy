import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { CreditCard, MoreHorizontal } from 'lucide-react'
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
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { platformKeyApi } from '@/api/keys'
import type { PlatformKey } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { EmptyState } from '@/components/ui/empty-state'
import { KeyStatusBadge } from '@/lib/label-badges'
import { useRowHighlight } from '@/lib/use-row-highlight'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export default function PlatformKeysPage() {
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const [keys, setKeys] = useState<PlatformKey[]>([])

  const load = useCallback(async () => {
    const res = await platformKeyApi.list()
    setKeys(res.items)
  }, [])

  useEffect(() => {
    let cancelled = false
    void platformKeyApi.list().then((res) => {
      if (!cancelled) setKeys(res.items)
    })
    return () => {
      cancelled = true
    }
  }, [])

  const handleRevoke = async (id: string) => {
    await platformKeyApi.revoke(id)
    toast.success('Key 已吊销')
    flashRow(id)
    void load()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <Button
          size="sm"
          className="bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-button"
          onClick={() =>
            open('key-create', {
              adminCreate: true,
              onSuccess: (id?: string) => {
                void load()
                if (id) flashRow(id)
              },
            })
          }
        >
          代建 Key
        </Button>
      </div>

      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          {keys.length === 0 ? (
            <EmptyState
              icon={CreditCard}
              title="暂无全局 Key"
              description="成员可在「我的 Key」中创建 Platform Key"
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    名称
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    绑定
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    Key 前缀
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    状态
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground w-36">
                    额度使用
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    模型白名单
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    到期时间
                  </TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((key) => {
                  const pct = Math.round((key.used / key.quota) * 100)
                  return (
                    <TableRow key={key.id} className={rowClass(key.id)}>
                      <TableCell className="font-medium">{key.name}</TableCell>
                      <TableCell className="text-sm">
                        {key.memberName ?? key.appName ?? '-'}
                      </TableCell>
                      <TableCell>
                        <span className="font-mono text-xs px-1.5 py-0.5 rounded bg-indigo-50 text-muted-foreground">
                          {key.keyPrefix}
                        </span>
                      </TableCell>
                      <TableCell>
                        <KeyStatusBadge status={key.status} />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Progress value={pct} className="flex-1 h-2" />
                          <span className="text-xs text-muted-foreground">{pct}%</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {key.modelWhitelist.slice(0, 2).map((m) => (
                            <Badge
                              key={m}
                              variant="outline"
                              className="text-xs bg-indigo-50 text-indigo-700 border-indigo-100"
                            >
                              {m}
                            </Badge>
                          ))}
                          {key.modelWhitelist.length > 2 && (
                            <Badge
                              variant="outline"
                              className="text-xs bg-indigo-50 text-indigo-700 border-indigo-100"
                            >
                              +{key.modelWhitelist.length - 2}
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {key.expiresAt ?? '永不'}
                      </TableCell>
                      <TableCell>
                        {key.status === 'active' && (
                          <DropdownMenu>
                            <DropdownMenuTrigger
                              render={
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              }
                            />
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                className="text-red-600"
                                onClick={() => handleRevoke(key.id)}
                              >
                                吊销
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
