import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageShell } from '@/components/layout/page-shell'
import { ConfirmActionDialog } from '@/components/ui/confirm-action-dialog'
import { cn } from '@/lib/utils'
import type { useMyKeysPage } from '@/features/keys'
import { MyKeysCardList } from './my-keys-card-list'

type MemberKeysPageShellProps = ReturnType<typeof useMyKeysPage>

export function MemberKeysPageShell({
  keys,
  budgetSummary,
  applyBudgetCta,
  createKeyCta,
  deleteTarget,
  setDeleteTarget,
  handleDelete,
  openCreateKey,
  openEditKey,
  openWithRefresh,
}: MemberKeysPageShellProps) {
  return (
    <PageShell
      description={<h1 className="text-sm font-semibold">我的 Key</h1>}
      actions={
        <>
          <Button
            id={applyBudgetCta.id}
            variant="outline"
            size="sm"
            className={cn(applyBudgetCta.className)}
            onClick={() => openWithRefresh('approval-submit', { defaultType: 'member_budget' })}
          >
            申请额度
          </Button>
          <Button
            id={createKeyCta.id}
            variant="default"
            size="sm"
            className={cn('gap-1.5', createKeyCta.className)}
            disabled={budgetSummary !== null && budgetSummary.remaining <= 0}
            onClick={() => openCreateKey()}
          >
            <Plus className="size-3.5" />
            新建 Key
          </Button>
        </>
      }
    >
      <div className="rounded-lg border border-border bg-card shadow-xs">
        <MyKeysCardList keys={keys} onEdit={openEditKey} onDelete={setDeleteTarget} />
      </div>

      <ConfirmActionDialog
        state={
          deleteTarget
            ? {
                open: true,
                title: '删除 Key？',
                desc: '删除后 Key 立即失效，不可恢复。已分配额度将释放回可用池。',
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
