import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { usePlatformKeysPage } from '@/features/keys'
import { useModelLabels } from '@/features/models/hooks/use-model-labels'
import { PlatformKeyTable } from './platform-key-table'
import { PlatformKeysDeptTree } from './platform-keys-dept-tree'
import { PlatformKeysToolbar } from './platform-keys-toolbar'

type PlatformKeysPageShellProps = ReturnType<typeof usePlatformKeysPage>

export function PlatformKeysPageShell({
  departments,
  selectedDeptId,
  setSelectedDeptId,
  activeTab,
  setActiveTab,
  treeSearch,
  setTreeSearch,
  search,
  setSearch,
  expanded,
  toggleExpand,
  keys,
  loading,
  error,
  refresh,
  rowClass,
  handleRevoke,
  openCreateKey,
}: PlatformKeysPageShellProps) {
  const { labelFor } = useModelLabels()
  return (
    <PageShell layout="fill">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        skeletonColumns={8}
        className="flex h-full min-h-0 flex-col overflow-hidden border-border shadow-xs"
        contentClassName="flex h-full min-h-0 flex-col p-0"
      >
        <div className="flex h-full min-h-0 overflow-hidden rounded-lg border border-border bg-card shadow-xs">
          <PlatformKeysDeptTree
            departments={departments}
            selectedId={selectedDeptId}
            onSelect={setSelectedDeptId}
            expanded={expanded}
            onToggle={toggleExpand}
            treeSearch={treeSearch}
            onTreeSearchChange={setTreeSearch}
          />

          <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
            <PlatformKeysToolbar
              activeTab={activeTab}
              onTabChange={setActiveTab}
              search={search}
              onSearchChange={setSearch}
              onCreateKey={openCreateKey}
            />
            <div className="flex-1 overflow-auto px-5 py-4">
              <PlatformKeyTable
                keys={keys}
                type={activeTab}
                rowClass={rowClass}
                onRevoke={handleRevoke}
                modelLabel={labelFor}
              />
            </div>
          </div>
        </div>
      </DataSection>
    </PageShell>
  )
}
