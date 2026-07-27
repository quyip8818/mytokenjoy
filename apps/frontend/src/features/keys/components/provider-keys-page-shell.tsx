import { Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import type { useProviderKeysPage } from '@/features/keys'
import { ProviderKeyTable } from './provider-key-table'

type ProviderKeysPageShellProps = ReturnType<typeof useProviderKeysPage>

export function ProviderKeysPageShell({
  keys,
  loading,
  error,
  refresh,
  rowClass,
  handleToggle,
  handleDelete,
  openForm,
}: ProviderKeysPageShellProps) {
  return (
    <PageShell>
      <PageHeader
        title="供应商 Key"
        actions={
          <PermissionGate write permission={PERMISSION.KEYS_PROVIDER}>
            <Button size="sm" variant="brand" onClick={() => openForm()}>
              添加 Provider Key
            </Button>
          </PermissionGate>
        }
      />

      <Card className="border-border shadow-xs">
        <CardContent className="px-5 pt-5 pb-4">
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
        </CardContent>
      </Card>
    </PageShell>
  )
}
