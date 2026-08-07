import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { DataSection } from '@/components/layout/data-section'
import type { usePlatformKeysPage } from '@/features/keys'
import { useModelLabels } from '@/features/models'
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
  handleDelete,
  openCreateKey,
}: PlatformKeysPageShellProps) {
  const { labelFor } = useModelLabels()
  return (
    <PageShell testId="page-keys-platform" className="flex min-h-0 flex-1 flex-col">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        skeletonColumns={8}
      >
        <SplitPanel
          master={
            <PlatformKeysDeptTree
              departments={departments}
              selectedId={selectedDeptId}
              onSelect={setSelectedDeptId}
              expanded={expanded}
              onToggle={toggleExpand}
              treeSearch={treeSearch}
              onTreeSearchChange={setTreeSearch}
            />
          }
          detail={
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
                  onDelete={handleDelete}
                  modelLabel={labelFor}
                />
              </div>
            </div>
          }
        />
      </DataSection>
    </PageShell>
  )
}
