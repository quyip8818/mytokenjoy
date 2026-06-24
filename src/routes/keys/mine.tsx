import { useCallback, useEffect, useState } from 'react'
import { MoreHorizontal, Plus, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import { platformKeyApi } from '@/api/keys'
import type { MemberQuotaSummary, PlatformKey } from '@/api/types'
import { useDemoRole, useDemoCta } from '@/features/demo'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { QUOTA_INSUFFICIENT_MESSAGE } from '@/features/workflow/constants'
import { EmptyState } from '@/components/ui/empty-state'
import { useRowHighlight } from '@/lib/use-row-highlight'
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
import { KeyStatusBadge } from '@/lib/label-badges'
import { Progress } from '@/components/ui/progress'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'

export default function MyKeysPage() {
  const { memberId } = useDemoRole()
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const applyQuotaCta = useDemoCta('APPLY_QUOTA')
  const createKeyCta = useDemoCta('CREATE_KEY')
  const [keys, setKeys] = useState<PlatformKey[]>([])
  const [quota, setQuota] = useState<MemberQuotaSummary | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<PlatformKey | null>(null)

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const [keyRes, quotaRes] = await Promise.all([
        platformKeyApi.list({ memberId }),
        platformKeyApi.getQuotaSummary(memberId),
      ])
      if (!cancelled) {
        setKeys(keyRes.items)
        setQuota(quotaRes)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [memberId])

  const refresh = useCallback(async () => {
    const [keyRes, quotaRes] = await Promise.all([
      platformKeyApi.list({ memberId }),
      platformKeyApi.getQuotaSummary(memberId),
    ])
    setKeys(keyRes.items)
    setQuota(quotaRes)
  }, [memberId])

  const handleDelete = async () => {
    if (!deleteTarget) return
    await platformKeyApi.delete(deleteTarget.id)
    toast.success('Key 已删除')
    setDeleteTarget(null)
    refresh()
  }

  const handleToggle = async (key: PlatformKey) => {
    const enabled = key.status !== 'active'
    await platformKeyApi.toggle(key.id, enabled)
    toast.success(enabled ? 'Key 已启用' : 'Key 已禁用')
    flashRow(key.id)
    refresh()
  }

  const openCreateKey = () => {
    if (quota !== null && quota.remaining <= 0) {
      toast.error(QUOTA_INSUFFICIENT_MESSAGE)
      return
    }
    open('key-create', {
      onSuccess: (id?: string) => {
        void refresh()
        if (id) flashRow(id)
      },
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end gap-3">
        <Button
          id={applyQuotaCta.id}
          variant="outline"
          className={cn(applyQuotaCta.className)}
          onClick={() => open('approval-submit', { defaultType: 'quota', onSuccess: refresh })}
        >
          申请额度
        </Button>
        <Button
          id={createKeyCta.id}
          className={cn(
            'bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-button hover:from-indigo-500 hover:to-violet-500',
            createKeyCta.className,
          )}
          disabled={quota !== null && quota.remaining <= 0}
          onClick={openCreateKey}
        >
          <Plus className="mr-1.5 h-4 w-4" />
          创建 Key
        </Button>
      </div>

      {quota && (
        <div className="grid grid-cols-3 gap-4">
          <Card className="border-border/50 shadow-card">
            <CardContent className="pt-5">
              <p className="text-xs text-muted-foreground">总额度</p>
              <p className="text-2xl font-bold">¥{quota.totalQuota.toLocaleString()}</p>
            </CardContent>
          </Card>
          <Card className="border-border/50 shadow-card">
            <CardContent className="pt-5">
              <p className="text-xs text-muted-foreground">已使用</p>
              <p className="text-2xl font-bold">¥{quota.used.toLocaleString()}</p>
            </CardContent>
          </Card>
          <Card className="border-border/50 shadow-card">
            <CardContent className="pt-5">
              <p className="text-xs text-muted-foreground">剩余</p>
              <p className="text-2xl font-bold text-indigo-600">
                ¥{quota.remaining.toLocaleString()}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      <Card className="border-border/50 shadow-card">
        <CardContent className="pt-5 pb-4">
          {keys.length === 0 ? (
            <EmptyState
              icon={KeyRound}
              title="还没有 Key"
              description="创建 Platform Key 后即可调用模型 API"
              actionLabel="创建第一个 Key"
              onAction={openCreateKey}
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    名称
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    Key 前缀
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    额度
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    模型
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    状态
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
                    <TableCell className="font-mono text-sm text-muted-foreground">
                      {key.keyPrefix}
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1 min-w-28">
                        <div className="text-xs text-muted-foreground">
                          ¥{key.used.toLocaleString()} / ¥{key.quota.toLocaleString()}
                        </div>
                        <Progress value={(key.used / key.quota) * 100} className="h-1.5" />
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {key.modelWhitelist.length} 个
                    </TableCell>
                    <TableCell>
                      <KeyStatusBadge status={key.status} />
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
                          <DropdownMenuItem
                            onClick={() => open('key-edit', { key, onSuccess: refresh })}
                          >
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() =>
                              open('key-rotate-confirm', {
                                key,
                                onRotate: (k: PlatformKey) => platformKeyApi.rotate(k.id),
                                onDone: refresh,
                              })
                            }
                          >
                            重新生成
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleToggle(key)}>
                            {key.status === 'active' ? '禁用' : '启用'}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-red-600"
                            onClick={() => setDeleteTarget(key)}
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

      <AlertDialog open={!!deleteTarget} onOpenChange={(v) => !v && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除 Key？</AlertDialogTitle>
            <AlertDialogDescription>此操作不可撤销。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-white">
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
