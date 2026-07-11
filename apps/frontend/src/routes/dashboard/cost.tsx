import {
  CostDashboardPageShell,
  useCostDashboardPage,
  useDeptSelection,
  useOrgTree,
  OrgTreeSidebar,
  DashboardPageLayout,
} from '@/features/dashboard'

export default function CostDashboardPage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useCostDashboardPage({ deptId: selectedDeptId })

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
        <h1 className="text-lg font-semibold">成本看板</h1>
      </div>
      <CostDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
