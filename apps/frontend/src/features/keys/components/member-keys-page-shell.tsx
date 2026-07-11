import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageShell } from '@/components/layout/page-shell'
import { cn } from '@/lib/utils'
import type { useMyKeysPage } from '@/features/keys'
import { MyKeysCardList } from './my-keys-card-list'

type MemberKeysPageShellProps = ReturnType<typeof useMyKeysPage>

export function MemberKeysPageShell({
  keys,
  quota,
  applyQuotaCta,
  createKeyCta,
  openCreateKey,
  openWithRefresh,
}: MemberKeysPageShellProps) {
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
            onClick={() => openWithRefresh('approval-submit', { defaultType: 'budget' })}
          >
            申请额度
          </Button>
          <Button
            id={createKeyCta.id}
            variant="default"
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
    >
      <div className="rounded-lg border border-border bg-card shadow-xs">
        <MyKeysCardList keys={keys} />
      </div>
    </PageShell>
  )
}
