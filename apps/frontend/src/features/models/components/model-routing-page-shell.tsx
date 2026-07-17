import { GitBranch } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
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
    <PageShell layout="fill" className="min-h-0 flex-1">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        className="flex min-h-0 flex-1 flex-col"
        contentClassName="flex min-h-0 flex-1 flex-col"
      >
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xs">
          <div className="flex min-h-0 flex-1 overflow-hidden">
            <RoutingTreePanel
              departments={departments}
              selectedId={selectedNodeId}
              onSelect={setSelectedNodeId}
            />
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden border-l border-border">
              {selectedRule && selectedDepartment ? (
                <PermissionGate
                  permission={PERMISSION.MODEL_WHITELIST}
                  fallback={
                    <div className="flex flex-1 items-center justify-center p-8">
                      <p className="text-sm text-muted-foreground">无权限配置模型</p>
                    </div>
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
                <div className="flex flex-1 flex-col items-center justify-center gap-3 p-8">
                  <GitBranch className="size-8 text-muted-foreground/50" />
                  <p className="text-sm text-muted-foreground">选择左侧团队查看模型配置</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </DataSection>
    </PageShell>
  )
}
