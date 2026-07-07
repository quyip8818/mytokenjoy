import type { ReactNode } from 'react'
import { KeyRound, Plus } from 'lucide-react'
import { listEmpty } from '@/lib/list-empty'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatCard } from '@/components/ui/stat-card'
import { Button } from '@/components/ui/button'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { cn } from '@/lib/utils'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useMyKeysPage } from '../hooks/use-my-keys-page'
import { MyKeysTable } from './my-keys-table'

type MyKeysPageShellProps = {
  memberPortal?: boolean
  description?: ReactNode
}

export function MyKeysPageShell({ memberPortal = false, description }: MyKeysPageShellProps) {
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
    handleToggleWithFlash,
    rowClass,
    openCreateKey,
    openEditKey,
    openRotateKey,
    openWithRefresh,
  } = useMyKeysPage()

  const applyQuotaButton = (
    <Button
      id={applyQuotaCta.id}
      variant="outline"
      size={memberPortal ? 'sm' : 'default'}
      className={cn(applyQuotaCta.className)}
      onClick={() => openWithRefresh('approval-submit', { defaultType: 'quota' })}
    >
      申请额度
    </Button>
  )

  const createKeyButton = (
    <Button
      id={createKeyCta.id}
      variant={memberPortal ? 'default' : 'brand'}
      size={memberPortal ? 'sm' : 'default'}
      className={cn(memberPortal ? 'gap-1.5' : undefined, createKeyCta.className)}
      disabled={quota !== null && quota.remaining <= 0}
      onClick={() => openCreateKey()}
    >
      <Plus className={memberPortal ? 'size-3.5' : 'mr-1.5 h-4 w-4'} />
      {memberPortal ? '新建 Key' : '创建 Key'}
    </Button>
  )

  return (
    <PageShell
      description={description}
      actions={
        memberPortal ? (
          <>
            {applyQuotaButton}
            {createKeyButton}
          </>
        ) : (
          <>
            <PermissionGate permission={PERMISSION.SELF_APPROVAL}>
              {applyQuotaButton}
            </PermissionGate>
            <PermissionGate write permission={PERMISSION.SELF_KEYS}>
              {createKeyButton}
            </PermissionGate>
          </>
        )
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
          memberPortal={memberPortal}
          onEdit={openEditKey}
          onRotate={openRotateKey}
          onToggle={handleToggleWithFlash}
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
