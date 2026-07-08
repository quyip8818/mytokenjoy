import { Link } from 'react-router'
import { ArrowRight, CheckCircle2, Settings2 } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { useDataSourcePage } from '@/features/org'
import { PLATFORM_LABELS } from '@/features/org'
import { Stepper } from './stepper'
import { PlatformSelect } from './platform-select'
import { PLATFORM_ICON_META } from './platform-meta'
import { StepCredentials } from './step-credentials'
import { StepFieldMapping } from './step-field-mapping'
import { StepSyncSchedule } from './step-sync-schedule'

const steps = [
  { title: '凭证配置', description: '连接第三方平台' },
  { title: '字段映射', description: '配置数据映射规则' },
  { title: '定时同步', description: '设置同步策略' },
]

type DataSourcePageShellProps = ReturnType<typeof useDataSourcePage>

export function DataSourcePageShell({
  phase,
  platform,
  currentStep,
  completedSteps,
  status,
  loading,
  error,
  refresh,
  dataSourceApi,
  syncApi,
  handlePlatformSelected,
  completeStep,
  handleWizardComplete,
  handleReconfigure,
  goToSelect,
  goToPreviousStep,
}: DataSourcePageShellProps) {
  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        className="border-0 bg-transparent py-0 shadow-none ring-0"
        contentClassName="p-0"
      >
        {phase === 'connected' && status?.connected && platform ? (
          <ConnectedCard
            platform={platform}
            lastImport={status.lastImport}
            onReconfigure={handleReconfigure}
          />
        ) : phase === 'select' ? (
          <div className="rounded-xl border border-border bg-card px-6 py-10 shadow-xs">
            <div className="mx-auto max-w-2xl">
              <PlatformSelect onSelect={handlePlatformSelected} />
            </div>
          </div>
        ) : (
          <div className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
            <div className="border-b border-border bg-muted/30 px-6 py-5">
              <Stepper
                steps={steps}
                currentStep={currentStep}
                completedSteps={completedSteps}
                onStepClick={goToPreviousStep}
              />
            </div>

            <div className="p-6">
              {currentStep === 0 && platform && (
                <StepCredentials
                  platform={platform}
                  dataSourceApi={dataSourceApi}
                  onConnected={() => completeStep(0)}
                  onBack={goToSelect}
                />
              )}
              {currentStep === 1 && platform && (
                <StepFieldMapping
                  platform={platform}
                  dataSourceApi={dataSourceApi}
                  onComplete={() => completeStep(1)}
                  onBack={() => goToPreviousStep(0)}
                />
              )}
              {currentStep === 2 && (
                <StepSyncSchedule
                  syncApi={syncApi}
                  onComplete={handleWizardComplete}
                  onBack={() => goToPreviousStep(1)}
                />
              )}
            </div>
          </div>
        )}
      </DataSection>
    </PageShell>
  )
}

function ConnectedCard({
  platform,
  lastImport,
  onReconfigure,
}: {
  platform: NonNullable<ReturnType<typeof useDataSourcePage>['platform']>
  lastImport?: string | null
  onReconfigure: () => void
}) {
  const meta = PLATFORM_ICON_META[platform]
  const Icon = meta.icon

  return (
    <div className="rounded-xl border border-border bg-card p-6 shadow-xs">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div
            className={`flex size-11 shrink-0 items-center justify-center rounded-lg ${meta.iconClassName}`}
          >
            <Icon className="size-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-semibold text-foreground">数据源已连接</h3>
              <Badge className="border-emerald-200 bg-emerald-50 text-emerald-700">
                <CheckCircle2 className="size-3" />
                {PLATFORM_LABELS[platform]}
              </Badge>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              {lastImport ? `上次导入：${lastImport}` : '尚未执行导入，可前往组织架构页导入数据'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onReconfigure}>
            <Settings2 className="size-3.5" />
            重新配置
          </Button>
          <Button size="sm" asChild>
            <Link to="/org/structure">
              前往组织架构
              <ArrowRight className="size-3.5" />
            </Link>
          </Button>
        </div>
      </div>
    </div>
  )
}
