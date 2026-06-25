import { Link } from 'react-router'
import { budgetApi } from '@/api/budget'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'

export default function BudgetAlertsPage() {
  const overrunCta = useDemoCta('OVERRUN')
  const {
    data: policy = null,
    loading,
    refresh,
  } = useAsyncResource(() => budgetApi.getOverrunPolicy(), [])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const notifyLabels = [
    policy?.notifyEmail && '邮箱',
    policy?.notifyPhone && '手机',
    policy?.notifyIm && 'IM',
  ].filter(Boolean)

  return (
    <PageShell>
      <DataSection
        title="全局超限策略"
        loading={loading}
        loadingVariant="spinner"
        headerAction={
          <PermissionGate write permission={PERMISSION.BUDGET_POLICY}>
            <Button
              id={overrunCta.id}
              size="sm"
              variant="brand"
              className={overrunCta.className}
              onClick={() => openWithRefresh('overrun-policy')}
            >
              编辑策略
            </Button>
          </PermissionGate>
        }
      >
        {policy && (
          <div className="space-y-4 text-sm">
            <div>
              <p className="mb-2 text-muted-foreground">预警阈值</p>
              <div className="flex gap-2">
                {policy.thresholds.map((t) => (
                  <Badge key={t} variant="outline">
                    {t}%
                  </Badge>
                ))}
              </div>
            </div>
            <div>
              <p className="mb-1 text-muted-foreground">通知渠道</p>
              <p>{notifyLabels.length > 0 ? notifyLabels.join('、') : '未配置'}</p>
            </div>
            <div>
              <p className="mb-1 text-muted-foreground">超限阻断文案</p>
              <p className="text-foreground/90">{policy.blockMessage}</p>
            </div>
            <p className="text-xs text-muted-foreground">
              超限行为固定为直接阻断。组织预算分配见{' '}
              <Link to="/budget/overview" className="text-blue-600 hover:underline">
                预算总览
              </Link>
              。
            </p>
          </div>
        )}
      </DataSection>
    </PageShell>
  )
}
