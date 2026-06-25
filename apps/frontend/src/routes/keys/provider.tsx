import { Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ProviderKeyTable } from '@/routes/keys/components/provider-key-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useProviderKeysPage } from '@/routes/keys/hooks/use-provider-keys-page'

export default function ProviderKeysPage() {
  const { keys, loading, error, refresh, rowClass, handleToggle, handleDelete, openForm } =
    useProviderKeysPage()

  return (
    <PageShell
      actions={
        <PermissionGate write permission={PERMISSION.KEYS_PROVIDER}>
          <Button size="sm" variant="brand" onClick={openForm}>
            添加 Key
          </Button>
        </PermissionGate>
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={7}
        empty={listEmpty(loading, keys, {
          icon: Key,
          title: '暂无供应商 Key',
          description: '添加供应商 Key 后可用于模型路由',
          actionLabel: '添加 Key',
          onAction: openForm,
        })}
      >
        <ProviderKeyTable
          keys={keys}
          rowClass={rowClass}
          onToggle={handleToggle}
          onDelete={handleDelete}
        />
      </DataSection>
    </PageShell>
  )
}
