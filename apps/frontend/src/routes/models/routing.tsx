import { GitBranch } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import { RoutingRulesTable } from '@/routes/models/components/routing-rules-table'
import { useModelRoutingPage } from '@/routes/models/hooks/use-model-routing-page'

export default function ModelRoutingPage() {
  const { rules, loading, error, refresh, getParentCount, openWhitelistConfig } =
    useModelRoutingPage()

  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={4}
        empty={listEmpty(loading, rules, {
          icon: GitBranch,
          title: '暂无路由规则',
          description: '组织节点将继承父级的模型白名单配置',
        })}
      >
        <RoutingRulesTable
          rules={rules}
          getParentCount={getParentCount}
          onConfigure={openWhitelistConfig}
        />
      </DataSection>
    </PageShell>
  )
}
