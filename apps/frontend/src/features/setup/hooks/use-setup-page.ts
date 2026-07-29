import { useCallback, useEffect, useState } from 'react'
import { setupApi, type SetupInitInput } from '@/api/setup'

type SetupState = 'loading' | 'ready' | 'submitting' | 'done' | 'not-setup'

export function useSetupPage() {
  const [state, setState] = useState<SetupState>('loading')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setupApi
      .status()
      .then((s) => {
        setState(s.ready ? 'ready' : 'not-setup')
      })
      .catch(() => {
        // Setup server not running (normal app is running) — redirect to login
        setState('not-setup')
      })
  }, [])

  const submit = useCallback(async (input: SetupInitInput) => {
    setState('submitting')
    setError(null)
    try {
      await setupApi.init(input)
      setState('done')
    } catch (err) {
      setError(err instanceof Error ? err.message : '初始化失败')
      setState('ready')
    }
  }, [])

  return { state, error, submit }
}
