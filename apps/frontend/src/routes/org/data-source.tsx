import type { Platform } from '@/api/types'
import { Stepper } from '@/components/org/data-source/stepper'
import { PlatformSelect } from '@/components/org/data-source/platform-select'
import { StepCredentials } from '@/components/org/data-source/step-credentials'
import { StepFieldMapping } from '@/components/org/data-source/step-field-mapping'
import { StepSyncSchedule } from '@/components/org/data-source/step-sync-schedule'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { CheckCircle2, Settings2 } from 'lucide-react'
import { useDataSourcePage } from './hooks/use-data-source-page'

const steps = [
  { title: '凭证配置', description: '连接第三方平台' },
  { title: '字段映射', description: '配置数据映射规则' },
  { title: '定时同步', description: '设置同步策略' },
]

const platformLabels: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export default function DataSourcePage() {
  const {
    phase,
    platform,
    currentStep,
    completedSteps,
    status,
    loading,
    handlePlatformSelected,
    completeStep,
    handleWizardComplete,
    handleReconfigure,
    goToSelect,
    goToPreviousStep,
  } = useDataSourcePage()

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-sm text-muted-foreground">加载中...</p>
      </div>
    )
  }

  if (phase === 'connected' && status?.connected && platform) {
    return (
      <div className="space-y-6">
        <div className="rounded-lg border border-border bg-card p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="flex size-8 items-center justify-center rounded-md bg-emerald-50">
                <CheckCircle2 className="size-4 text-emerald-600" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-semibold text-foreground">数据源已连接</h3>
                  <Badge className="bg-emerald-50 text-emerald-700">{platformLabels[platform]}</Badge>
                </div>
                <p className="text-xs text-muted-foreground mt-0.5">
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
      </div>
    )
  }

  if (phase === 'select') {
    return (
      <div className="mx-auto max-w-2xl py-8">
        <PlatformSelect onSelect={handlePlatformSelected} />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-border bg-white p-5 shadow-xs">
        <Stepper steps={steps} currentStep={currentStep} completedSteps={completedSteps} />
      </div>

      <div className="rounded-lg border border-border bg-card p-6 shadow-xs">
        {currentStep === 0 && platform && (
          <StepCredentials
            platform={platform}
            onConnected={() => completeStep(0)}
            onBack={goToSelect}
          />
        )}
        {currentStep === 1 && platform && (
          <StepFieldMapping
            platform={platform}
            onComplete={() => completeStep(1)}
            onBack={() => goToPreviousStep(0)}
          />
        )}
        {currentStep === 2 && (
          <StepSyncSchedule
            onComplete={handleWizardComplete}
            onBack={() => goToPreviousStep(1)}
          />
        )}
      </div>
    </div>
  )
}
