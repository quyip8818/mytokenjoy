import { toast } from 'sonner'
import { CreditCard } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PlatformKeyTable } from '@/components/keys/platform-key-table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { platformKeyApi } from '@/api/keys'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/lib/use-row-highlight'

export default function PlatformKeysPage() {
  const { flashRow, rowClass } = useRowHighlight()
  const {
    data: keys = [],
    loading,
    refresh,
  } = useAsyncResource(async () => {
    const res = await platformKeyApi.list()
    return res.items
  }, [])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

  const handleRevoke = async (id: string) => {
    await platformKeyApi.revoke(id)
    toast.success('Key 已吊销')
    flashRow(id)
    void refresh()
  }

  const openCreateKey = () => openWithRefresh('key-create', { adminCreate: true })

  return (
    <PageShell
      actions={
        <Button size="sm" variant="brand" onClick={openCreateKey}>
          代建 Key
        </Button>
      }
    >
      <DataSection
        loading={loading}
        skeletonColumns={8}
        empty={listEmpty(loading, keys, {
          icon: CreditCard,
          title: '暂无全局 Key',
          description: '成员可在「我的 Key」中创建 Platform Key，或由管理员代建',
          actionLabel: '代建 Key',
          onAction: openCreateKey,
        })}
      >
        <PlatformKeyTable keys={keys} rowClass={rowClass} onRevoke={handleRevoke} />
      </DataSection>
    </PageShell>
  )
}
