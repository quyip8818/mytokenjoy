import {
  UsageDashboardPageShell,
  useUsageDashboardPage,
  useDeptSelection,
  useOrgTree,
  OrgTreeSidebar,
  DashboardPageLayout,
} from '@/features/dashboard'

export default function UsageDashboardPage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useUsageDashboardPage({ deptId: selectedDeptId })

  return (
    <DashboardPageLayout
      sidebar={
        <OrgTreeSidebar
          departments={departments}
          selectedDeptId={selectedDeptId}
          onSelect={setSelectedDeptId}
          loading={treeLoading}
        />
      }
    >
      <div className="mb-4">
        <p className="text-xs text-muted-foreground">
          {getBreadcrumb(selectedDeptId).join(' > ')}
        </p>
        <h1 className="text-lg font-semibold">用量分析</h1>
      </div>
      <UsageDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
