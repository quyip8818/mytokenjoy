import { useState } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { Platform } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

export type DataSourceWizardPhase = 'loading' | 'select' | 'steps' | 'connected'

export function useDataSourcePage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [userPhase, setUserPhase] = useState<DataSourceWizardPhase | null>(null)
  const [userPlatform, setUserPlatform] = useState<Platform | null>(null)
  const [currentStep, setCurrentStep] = useState(0)
  const [completedSteps, setCompletedSteps] = useState<number[]>([])

  const {
    data: status,
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.dataSource(),
    queryFn: (a) => a.dataSourceApi.getStatus(),
  })

  const phase: DataSourceWizardPhase =
    userPhase ??
    (loading || status === undefined
      ? 'loading'
      : status.connected && status.platform
        ? 'connected'
        : 'select')
  const platform = userPlatform ?? (status?.connected && status.platform ? status.platform : null)

  const handlePlatformSelected = (p: Platform) => {
    setUserPlatform(p)
    setCurrentStep(0)
    setCompletedSteps([])
    setUserPhase('steps')
  }

  const completeStep = (step: number) => {
    setCompletedSteps((prev) => (prev.includes(step) ? prev : [...prev, step]))
    if (step < 2) setCurrentStep(step + 1)
  }

  const handleWizardComplete = async () => {
    setCompletedSteps([0, 1, 2])
    setUserPhase('connected')
    try {
      const result = await apis.dataSourceApi.import()
      const successMsg = `导入完成：${result.successMembers} 名成员，${result.successDepartments} 个部门`
      if (result.failures.length === 0) {
        toast.success(successMsg)
      } else {
        toast.success(successMsg)
        const reasons = [...new Set(result.failures.map((f) => f.reason))]
        toast.warning(
          `${result.failures.length} 名成员导入失败：${reasons.slice(0, 3).join('；')}`,
          { duration: 8000 },
        )
      }
    } catch {
      toast.info('配置已保存，可前往组织架构页执行导入')
    }
    await refresh()
  }

  const handleReconfigure = () => {
    setUserPhase('select')
    setCurrentStep(0)
    setCompletedSteps([])
  }

  const goToSelect = () => setUserPhase('select')
  const goToPreviousStep = (step: number) => setCurrentStep(step)

  return {
    phase,
    platform,
    currentStep,
    completedSteps,
    status: status ?? null,
    loading: phase === 'loading' && loading,
    error,
    refresh,
    dataSourceApi: apis.dataSourceApi,
    syncApi: apis.syncApi,
    handlePlatformSelected,
    completeStep,
    handleWizardComplete,
    handleReconfigure,
    goToSelect,
    goToPreviousStep,
  }
}
