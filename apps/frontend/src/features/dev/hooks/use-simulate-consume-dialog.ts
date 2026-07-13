import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import type { CallLog, PlatformKey } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys } from '@/features/query'
import { toast } from 'sonner'
import {
  DEFAULT_INPUT_TOKENS,
  DEFAULT_OUTPUT_TOKENS,
  PLATFORM_KEY_ID_SESSION_KEY,
  formatEstimatedConsume,
} from '../lib/constants'
import {
  GatewayClientError,
  fetchBaselineCallIds,
  pollForNewCall,
  postChatCompletions,
} from '../lib/simulate-consume'

type Phase = 'idle' | 'calling' | 'waiting'

function readStoredPlatformKeyId(): string {
  if (!import.meta.env.DEV) return ''
  return sessionStorage.getItem(PLATFORM_KEY_ID_SESSION_KEY) ?? ''
}

function writeStoredPlatformKeyId(platformKeyId: string) {
  if (!import.meta.env.DEV) return
  if (platformKeyId) sessionStorage.setItem(PLATFORM_KEY_ID_SESSION_KEY, platformKeyId)
  else sessionStorage.removeItem(PLATFORM_KEY_ID_SESSION_KEY)
}

function parseInputTokens(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10)
  if (!Number.isFinite(parsed)) return fallback
  return parsed
}

function parseOutputTokens(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback
}

function formatPlatformKeyLabel(key: PlatformKey): string {
  const prefix = key.keyPrefix ? ` · ${key.keyPrefix}` : ''
  return `${key.name}${prefix}`
}

async function resolveBearer(
  devApiClient: AppApis['devApi'],
  platformKeyId: string,
): Promise<string> {
  const { bearer } = await devApiClient.getPlatformKeyBearer(platformKeyId)
  if (!bearer) {
    throw new Error('未返回 sk-，请检查 Key 是否已同步到 NewAPI')
  }
  return bearer
}

export function useSimulateConsumeDialog(open: boolean, injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const [platformKeys, setPlatformKeys] = useState<PlatformKey[]>([])
  const [keysLoading, setKeysLoading] = useState(false)
  const [selectedKeyId, setSelectedKeyIdState] = useState(readStoredPlatformKeyId)
  const [bearer, setBearer] = useState('')
  const [resolvingKey, setResolvingKey] = useState(false)
  const [inputTokensText, setInputTokensText] = useState(String(DEFAULT_INPUT_TOKENS))
  const [outputTokensText, setOutputTokensText] = useState(String(DEFAULT_OUTPUT_TOKENS))
  const [phase, setPhase] = useState<Phase>('idle')
  const [error, setError] = useState<string | null>(null)
  const [matchedCall, setMatchedCall] = useState<CallLog | null>(null)

  const inputTokens = useMemo(
    () => parseInputTokens(inputTokensText, DEFAULT_INPUT_TOKENS),
    [inputTokensText],
  )
  const outputTokens = useMemo(
    () => parseOutputTokens(outputTokensText, DEFAULT_OUTPUT_TOKENS),
    [outputTokensText],
  )
  const estimatedCost = useMemo(
    () => formatEstimatedConsume(inputTokens, outputTokens),
    [inputTokens, outputTokens],
  )

  const platformKeyOptions = useMemo(
    () => Object.fromEntries(platformKeys.map((key) => [key.id, formatPlatformKeyLabel(key)])),
    [platformKeys],
  )

  const selectPlatformKey = useCallback(
    async (platformKeyId: string) => {
      setSelectedKeyIdState(platformKeyId)
      writeStoredPlatformKeyId(platformKeyId)
      if (!platformKeyId) {
        setBearer('')
        return
      }
      setResolvingKey(true)
      setError(null)
      try {
        const sk = await resolveBearer(apis.devApi, platformKeyId)
        setBearer(sk)
      } catch (err) {
        setBearer('')
        setError(err instanceof Error ? err.message : '无法获取 Platform Key')
      } finally {
        setResolvingKey(false)
      }
    },
    [apis.devApi],
  )

  useEffect(() => {
    if (!open) return
    let cancelled = false
    setKeysLoading(true)
    void apis.platformKeyApi
      .list({ scope: 'member', page: 1, pageSize: 50 })
      .then((page) => {
        if (cancelled) return
        const activeKeys = page.items.filter((key) => key.status === 'active')
        setPlatformKeys(activeKeys)

        const storedId = readStoredPlatformKeyId()
        const initialId =
          activeKeys.find((key) => key.id === storedId)?.id ?? activeKeys[0]?.id ?? ''
        if (initialId) {
          void selectPlatformKey(initialId)
        } else {
          setSelectedKeyIdState('')
          setBearer('')
        }
      })
      .catch(() => {
        if (!cancelled) {
          setPlatformKeys([])
          setError('无法加载 Platform Key 列表')
        }
      })
      .finally(() => {
        if (!cancelled) setKeysLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [apis.platformKeyApi, open, selectPlatformKey])

  const handleSubmit = useCallback(async () => {
    if (!selectedKeyId) {
      setError('请选择 Platform Key')
      return
    }
    if (inputTokens < 1) {
      setError('Input tokens 须 ≥ 1')
      return
    }

    let sk = bearer.trim()
    if (!sk) {
      try {
        sk = await resolveBearer(apis.devApi, selectedKeyId)
        setBearer(sk)
      } catch (err) {
        setError(err instanceof Error ? err.message : '无法获取 Platform Key')
        return
      }
    }

    setPhase('calling')
    setError(null)
    setMatchedCall(null)

    try {
      const baselineIds = await fetchBaselineCallIds(apis.auditApi)
      await postChatCompletions({ bearer: sk, inputTokens, outputTokens })
      toast.message('已调用，等待入账…')
      setPhase('waiting')

      const call = await pollForNewCall(apis.auditApi, baselineIds, inputTokens, outputTokens)
      if (!call) {
        setError('轮询超时：请检查 Ingest Worker 是否在运行')
        return
      }

      setMatchedCall(call)
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.audit.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.wallet.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.budget.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all }),
      ])
      toast.success('入账完成，可在调用审计中查看')
    } catch (err) {
      if (err instanceof GatewayClientError) {
        setError(err.body || `Gateway 预检失败 (${err.status})`)
      } else {
        setError(err instanceof Error ? err.message : '提交失败')
      }
    } finally {
      setPhase('idle')
    }
  }, [apis.auditApi, apis.devApi, bearer, inputTokens, outputTokens, queryClient, selectedKeyId])

  return {
    platformKeys,
    platformKeyOptions,
    selectedKeyId,
    setSelectedKeyId: selectPlatformKey,
    keysLoading,
    resolvingKey,
    inputTokensText,
    setInputTokensText,
    outputTokensText,
    setOutputTokensText,
    estimatedCost,
    busy: phase !== 'idle' || resolvingKey,
    waiting: phase === 'waiting',
    error,
    matchedCall,
    handleSubmit,
  }
}
