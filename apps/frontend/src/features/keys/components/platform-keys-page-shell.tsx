import { CreditCard } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import type { usePlatformKeysPage } from '@/features/keys/hooks/use-platform-keys-page'
import { PlatformKeyTable } from './platform-key-table'

type PlatformKeysPageShellProps = ReturnType<typeof usePlatformKeysPage>

export function PlatformKeysPageShell({
  keys,
  loading,
  error,
  refresh,
  rowClass,
  handleRevoke,
  openCreateKey,
}: PlatformKeysPageShellProps) {
  return (
    <PageShell
      actions={
        <PermissionGate write permission={PERMISSION.KEYS_ADMIN}>
          <Button size="sm" variant="brand" onClick={() => openCreateKey()}>
            代建 Key
          </Button>
        </PermissionGate>
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={8}
        empty={listEmpty(loading, keys, {
          icon: CreditCard,
          title: '暂无全局 Key',
          description: '成员可在「我的 Key」中创建 Platform Key，或由管理员代建',
          actionLabel: '代建 Key',
          onAction: () => openCreateKey(),
        })}
      >
        <PlatformKeyTable keys={keys} rowClass={rowClass} onRevoke={handleRevoke} />
      </DataSection>
    </PageShell>
  )
}
