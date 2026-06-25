import { KeyRound, Plus } from 'lucide-react'
import type { PlatformKey } from '@/api/types'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/lib/use-row-highlight'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatCard } from '@/components/ui/stat-card'
import { Button } from '@/components/ui/button'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { cn } from '@/lib/utils'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useMyKeysPage } from '@/routes/keys/hooks/use-my-keys-page'
import { MyKeysTable } from '@/routes/keys/components/my-keys-table'

export default function MyKeysPage() {
  const { flashRow, rowClass } = useRowHighlight()
  const {
    keys,
    quota,
    loading,
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
              className={cn(createKeyCta.className)}
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
        <MyKeysTable
          keys={keys}
          rowClass={rowClass}
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
