import { useState } from 'react'
import { MoreHorizontal, Plus, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import { platformKeyApi } from '@/api/keys'
import type { PlatformKey } from '@/api/types'
import { useDemoRole, useDemoCta } from '@/features/demo'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { QUOTA_INSUFFICIENT_MESSAGE } from '@/features/workflow/constants'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/lib/use-row-highlight'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatCard } from '@/components/ui/stat-card'
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
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'

export default function MyKeysPage() {
  const { memberId } = useDemoRole()
  const { flashRow, rowClass } = useRowHighlight()
  const applyQuotaCta = useDemoCta('APPLY_QUOTA')
  const createKeyCta = useDemoCta('CREATE_KEY')
  const [deleteTarget, setDeleteTarget] = useState<PlatformKey | null>(null)

  const { data, loading, refresh } = useAsyncResource(async () => {
    const [keyRes, quotaRes] = await Promise.all([
      platformKeyApi.list({ memberId }),
      platformKeyApi.getQuotaSummary(memberId),
    ])
    return { keys: keyRes.items, quota: quotaRes }
  }, [memberId])
  const keys = data?.keys ?? []
  const quota = data?.quota ?? null
  const { openWithRefresh, open } = useWorkflowRefresh(refresh, flashRow)

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
    openWithRefresh('key-create')
  }

  return (
    <PageShell
      actions={
        <>
          <PermissionGate permission={PERMISSION.SELF_APPROVAL}>
            <Button
              id={applyQuotaCta.id}
              variant="outline"
              className={cn(applyQuotaCta.className)}
              onClick={() => openWithRefresh('approval-submit', { defaultType: 'quota' })}
            >
              申请额度
            </Button>
          </PermissionGate>
          <PermissionGate write permission={PERMISSION.SELF_KEYS}>
            <Button
              id={createKeyCta.id}
              variant="brand"
              className={createKeyCta.className}
              disabled={quota !== null && quota.remaining <= 0}
              onClick={openCreateKey}
            >
              <Plus className="mr-1.5 h-4 w-4" />
              创建 Key
            </Button>
          </PermissionGate>
        </>
      }
      stats={
        quota ? (
          <div className="grid grid-cols-3 gap-4">
            <StatCard label="总额度" value={`¥${quota.totalQuota.toLocaleString()}`} />
            <StatCard label="已使用" value={`¥${quota.used.toLocaleString()}`} />
            <StatCard label="剩余" value={`¥${quota.remaining.toLocaleString()}`} accent />
          </div>
        ) : loading ? (
          <div className="grid grid-cols-3 gap-4">
            <StatCard label="总额度" value="-" />
            <StatCard label="已使用" value="-" />
            <StatCard label="剩余" value="-" />
          </div>
        ) : null
      }
    >
      <DataSection
        loading={loading}
        skeletonColumns={6}
        empty={listEmpty(loading, keys, {
          icon: KeyRound,
          title: '还没有 Key',
          description: '创建 Platform Key 后即可调用模型 API',
          actionLabel: '创建第一个 Key',
          onAction: openCreateKey,
        })}
      >
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>名称</TableHead>
              <TableHead>Key 前缀</TableHead>
              <TableHead>额度</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="w-[120px]">操作</TableHead>
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
                  <div className="min-w-28 space-y-1">
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
                  <PermissionGate write permission={PERMISSION.SELF_KEYS}>
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
                  </PermissionGate>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>

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
    </PageShell>
  )
}
