import { useEffect, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { Platform } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

export type DataSourceWizardPhase = 'loading' | 'select' | 'steps' | 'connected'

export function useDataSourcePage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [phase, setPhase] = useState<DataSourceWizardPhase>('loading')
  const [platform, setPlatform] = useState<Platform | null>(null)
  const [currentStep, setCurrentStep] = useState(0)
  const [completedSteps, setCompletedSteps] = useState<number[]>([])

  const { data: status, loading, refresh } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.org.dataSource(),
    queryFn: (a) => a.dataSourceApi.getStatus(),
  })

  useEffect(() => {
    if (loading || status === undefined) return
    if (phase !== 'loading') return
    if (status.connected && status.platform) {
      setPlatform(status.platform)
      setPhase('connected')
    } else {
      setPhase('select')
    }
  }, [loading, status, phase])

  const handlePlatformSelected = (p: Platform) => {
    setPlatform(p)
    setCurrentStep(0)
    setCompletedSteps([])
    setPhase('steps')
  }

  const completeStep = (step: number) => {
    setCompletedSteps((prev) => (prev.includes(step) ? prev : [...prev, step]))
    if (step < 2) setCurrentStep(step + 1)
  }

  const handleWizardComplete = async () => {
    setCompletedSteps([0, 1, 2])
    setPhase('connected')
    await refresh()
  }

  const handleReconfigure = () => {
    setPhase('select')
    setCurrentStep(0)
    setCompletedSteps([])
  }

  const goToSelect = () => setPhase('select')
  const goToPreviousStep = (step: number) => setCurrentStep(step)

  return {
    phase,
    platform,
    currentStep,
    completedSteps,
    status: status ?? null,
    loading: phase === 'loading' && loading,
    handlePlatformSelected,
    completeStep,
    handleWizardComplete,
    handleReconfigure,
    goToSelect,
    goToPreviousStep,
  }
}
