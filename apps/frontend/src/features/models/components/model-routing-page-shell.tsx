import { GitBranch } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { listEmpty } from '@/lib/list-empty'
import type { useModelRoutingPage } from '@/features/models/hooks/use-model-routing-page'
import { RoutingRulesTable } from './routing-rules-table'

type ModelRoutingPageShellProps = ReturnType<typeof useModelRoutingPage>

export function ModelRoutingPageShell({
  rules,
  loading,
  error,
  refresh,
  getParentCount,
  openWhitelistConfig,
}: ModelRoutingPageShellProps) {
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
