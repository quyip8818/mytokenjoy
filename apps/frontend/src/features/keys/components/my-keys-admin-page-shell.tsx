import { KeyRound } from 'lucide-react'
import { listEmpty } from '@/lib/list-empty'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatCard } from '@/components/ui/stat-card'
import { Button } from '@/components/ui/button'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import type { useMyKeysPage } from '@/features/keys'
import { MyKeysTable } from './my-keys-table'

type MyKeysAdminPageShellProps = ReturnType<typeof useMyKeysPage>

export function MyKeysAdminPageShell({
  keys,
  budgetSummary,
  loading,
  error,
  refresh,
  deleteTarget,
  setDeleteTarget,
  applyBudgetCta,
  createKeyCta,
  handleDelete,
  handleToggleWithFlash,
  rowClass,
  openCreateKey,
  openEditKey,
  openRotateKey,
  openWithRefresh,
}: MyKeysAdminPageShellProps) {
  const applyBudgetButton = (
    <Button
      id={applyBudgetCta.id}
      variant="outline"
      className={applyBudgetCta.className}
      onClick={() => openWithRefresh('approval-submit', { defaultType: 'budget' })}
    >
      申请额度
    </Button>
  )

  const createKeyButton = (
    <Button
      id={createKeyCta.id}
      variant="brand"
      className={createKeyCta.className}
      disabled={budgetSummary !== null && budgetSummary.remaining <= 0}
      onClick={() => openCreateKey()}
    >
      创建 Key
    </Button>
  )

  return (
    <PageShell
      actions={
        <>
          <PermissionGate permission={PERMISSION.SELF_APPROVAL}>{applyBudgetButton}</PermissionGate>
          <PermissionGate write permission={PERMISSION.SELF_KEYS}>
            {createKeyButton}
          </PermissionGate>
        </>
      }
      stats={
        budgetSummary ? (
          <div className="grid grid-cols-3 gap-4">
            <StatCard label="总额度" value={`¥${budgetSummary.totalBudget.toLocaleString()}`} />
            <StatCard label="已使用" value={`¥${budgetSummary.consumed.toLocaleString()}`} />
            <StatCard label="剩余" value={`¥${budgetSummary.remaining.toLocaleString()}`} accent />
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
