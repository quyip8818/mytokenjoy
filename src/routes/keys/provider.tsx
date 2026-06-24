import { toast } from 'sonner'
import { Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ProviderKeyTable } from '@/components/keys/provider-key-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { providerKeyApi } from '@/api/keys'
import type { ProviderKey } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/lib/use-row-highlight'

export default function ProviderKeysPage() {
  const { flashRow, rowClass } = useRowHighlight()
  const { data: keys = [], loading, refresh } = useAsyncResource(() => providerKeyApi.list(), [])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

  const handleToggle = async (key: ProviderKey) => {
    const enabled = key.status !== 'active'
    await providerKeyApi.toggle(key.id, enabled)
    toast.success(enabled ? 'Key 已启用' : 'Key 已禁用')
    flashRow(key.id)
    void refresh()
  }

  const handleDelete = async (id: string) => {
    await providerKeyApi.delete(id)
    toast.success('Key 已删除')
    void refresh()
  }

  const openForm = () => openWithRefresh('provider-key-form')

  return (
    <PageShell
      actions={
        <Button size="sm" variant="brand" onClick={openForm}>
          添加 Key
        </Button>
      }
    >
      <DataSection
        loading={loading}
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
