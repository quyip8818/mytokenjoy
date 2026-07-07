import { CheckCircle2, Settings2 } from 'lucide-react'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { useDataSourcePage } from '@/features/org'
import { PLATFORM_LABELS } from '@/features/org'
import { Stepper } from './stepper'
import { PlatformSelect } from './platform-select'
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
    <PageShell className="space-y-6">
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6"
      >
        {phase === 'connected' && status?.connected && platform ? (
          <div className="rounded-lg border border-border bg-card p-5 shadow-xs">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex size-8 items-center justify-center rounded-md bg-emerald-50">
                  <CheckCircle2 className="size-4 text-emerald-600" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold text-foreground">数据源已连接</h3>
                    <Badge className="bg-emerald-50 text-emerald-700">
                      {PLATFORM_LABELS[platform]}
                    </Badge>
                  </div>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {status.lastImport ? `上次导入：${status.lastImport}` : '尚未执行导入'}
                  </p>
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={handleReconfigure}>
                <Settings2 className="size-3.5" />
                重新配置
              </Button>
            </div>
          </div>
        ) : phase === 'select' ? (
          <div className="mx-auto max-w-2xl py-8">
            <PlatformSelect onSelect={handlePlatformSelected} />
          </div>
        ) : (
          <>
            <div className="rounded-lg border border-border bg-white p-5 shadow-xs">
              <Stepper steps={steps} currentStep={currentStep} completedSteps={completedSteps} />
            </div>

            <div className="rounded-lg border border-border bg-card p-6 shadow-xs">
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
          </>
        )}
      </DataSection>
    </PageShell>
  )
}
