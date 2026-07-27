import { GitBranch } from 'lucide-react'
import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { DataSection } from '@/components/layout/data-section'
import { EmptyState } from '@/components/ui/empty-state'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import type { useModelRoutingPage } from '@/features/models'
import { RoutingDetailPanel } from './routing-detail-panel'
import { RoutingTreePanel } from './routing-tree-panel'

type ModelRoutingPageShellProps = ReturnType<typeof useModelRoutingPage>

export function ModelRoutingPageShell({
  departments,
  models,
  selectedNodeId,
  setSelectedNodeId,
  selectedRule,
  selectedDepartment,
  parentRule,
  loading,
  error,
  refresh,
  saving,
  handleSave,
}: ModelRoutingPageShellProps) {
  return (
    <PageShell className="flex min-h-0 flex-1 flex-col">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        loadingVariant="spinner"
      >
        <SplitPanel
          master={
            <RoutingTreePanel
              departments={departments}
              selectedId={selectedNodeId}
              onSelect={setSelectedNodeId}
            />
          }
          detail={
            selectedRule && selectedDepartment ? (
              <PermissionGate
                permission={PERMISSION.MODEL_WHITELIST}
                fallback={
                  <EmptyState
                    variant="minimal"
                    icon={GitBranch}
                    title="无权限配置模型"
                  />
                }
              >
                <RoutingDetailPanel
                  department={selectedDepartment}
                  rule={selectedRule}
                  parentRule={parentRule}
                  models={models}
                  saving={saving}
                  onSave={handleSave}
                />
              </PermissionGate>
            ) : (
              <EmptyState
                variant="minimal"
                icon={GitBranch}
                title="选择左侧团队查看模型配置"
              />
            )
          }
        />
      </DataSection>
    </PageShell>
  )
}
