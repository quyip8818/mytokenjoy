import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { ROUTES } from '@/config/routes'
import { PERMISSION } from '@/lib/permissions'
import { usePermissions } from '@/hooks/use-permissions'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import {
  MemberConsumptionCharts,
  MemberDashboardStats,
  useMemberDashboardPage,
} from '@/features/member'

export default function MemberDashboardPage() {
  const navigate = useNavigate()
  const { has } = usePermissions()
  const {
    loading,
    error,
    refresh,
    accountData,
    usageStats,
    resourceConsumption,
    performance,
    consumptionTrend,
    consumptionDistribution,
    callDistribution,
    callRanking,
    distributionTotal,
    trendTotal,
    callTotal,
  } = useMemberDashboardPage()

  const handleRecharge = () => {
    if (has([PERMISSION.BILLING_READ, PERMISSION.BILLING_RECHARGE])) {
      navigate(ROUTES.wallet)
      return
    }
    toast.message('请联系管理员进行充值')
  }

  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6"
      >
        <MemberDashboardStats
          loading={loading}
          accountData={accountData}
          usageStats={usageStats}
          resourceConsumption={resourceConsumption}
          performance={performance}
          onRecharge={handleRecharge}
        />
        <MemberConsumptionCharts
          loading={loading}
          consumptionDistribution={consumptionDistribution}
          consumptionTrend={consumptionTrend}
          callDistribution={callDistribution}
          callRanking={callRanking}
          distributionTotal={distributionTotal}
          trendTotal={trendTotal}
          callTotal={callTotal}
        />
      </DataSection>
    </PageShell>
  )
}
