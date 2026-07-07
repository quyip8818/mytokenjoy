import { Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { ProviderKeyTable, useProviderKeysPage } from '@/features/keys'

export default function ProviderKeysPage() {
  const { keys, loading, error, refresh, rowClass, handleToggle, handleDelete, openForm } =
    useProviderKeysPage()

  return (
    <PageShell
      actions={
        <PermissionGate write permission={PERMISSION.KEYS_PROVIDER}>
          <Button size="sm" variant="brand" onClick={() => openForm()}>
            添加 Provider Key
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
          title: '暂无 Provider Key',
          description: '添加 Provider Key 以接入外部模型服务',
          actionLabel: '添加 Provider Key',
          onAction: () => openForm(),
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
