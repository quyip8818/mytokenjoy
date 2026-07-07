import { KeyRound, Plus } from 'lucide-react'
import type { PlatformKey } from '@/api/types'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatCard } from '@/components/ui/stat-card'
import { Button } from '@/components/ui/button'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { cn } from '@/lib/utils'
import { useMyKeysPage, MyKeysTable } from '@/features/keys'

export default function MemberKeysPage() {
  const { flashRow, rowClass } = useRowHighlight()
  const {
    keys,
    quota,
    loading,
    error,
    refresh,
    deleteTarget,
    setDeleteTarget,
    applyQuotaCta,
    createKeyCta,
    handleDelete,
    handleToggle,
    openCreateKey,
    openEditKey,
    openRotateKey,
    openWithRefresh,
  } = useMyKeysPage()

  const onToggle = async (key: PlatformKey) => {
    const id = await handleToggle(key)
    flashRow(id)
  }

  return (
    <PageShell
      description={<h1 className="text-sm font-semibold">我的 Key</h1>}
      actions={
        <>
          <Button
            id={applyQuotaCta.id}
            variant="outline"
            size="sm"
            className={cn(applyQuotaCta.className)}
            onClick={() => openWithRefresh('approval-submit', { defaultType: 'quota' })}
          >
            申请额度
          </Button>
          <Button
            id={createKeyCta.id}
            size="sm"
            className={cn('gap-1.5', createKeyCta.className)}
            disabled={quota !== null && quota.remaining <= 0}
            onClick={() => openCreateKey()}
          >
            <Plus className="size-3.5" />
            新建 Key
          </Button>
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
        error={error}
        onRetry={() => void refresh()}
        skeletonColumns={6}
        empty={listEmpty(loading, keys, {
          icon: KeyRound,
          title: '还没有 Key',
          description: '创建 Platform Key 后即可调用模型 API',
          actionLabel: '创建第一个 Key',
          onAction: () => openCreateKey(),
        })}
      >
        <MyKeysTable
          keys={keys}
          rowClass={rowClass}
          memberPortal
          onEdit={openEditKey}
          onRotate={openRotateKey}
          onToggle={onToggle}
          onDelete={setDeleteTarget}
        />
      </DataSection>

      <ConfirmActionDialog
        state={
          deleteTarget
            ? {
                open: true,
                title: '删除 Key？',
                desc: '此操作不可撤销。',
                variant: 'danger',
                confirmLabel: '删除',
                onConfirm: handleDelete,
              }
            : null
        }
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        onClose={() => setDeleteTarget(null)}
      />
    </PageShell>
  )
}
