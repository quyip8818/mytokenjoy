import { useCallback, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import type { CallLog, PlatformKey } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
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

  const [userSelectedKeyId, setUserSelectedKeyId] = useState(readStoredPlatformKeyId)
  const [inputTokensText, setInputTokensText] = useState(String(DEFAULT_INPUT_TOKENS))
  const [outputTokensText, setOutputTokensText] = useState(String(DEFAULT_OUTPUT_TOKENS))
  const [phase, setPhase] = useState<Phase>('idle')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [matchedCall, setMatchedCall] = useState<CallLog | null>(null)

  const {
    data: platformKeys = [],
    loading: keysLoading,
    error: keysQueryError,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.keys.all, 'simulate-consume'] as const,
    queryFn: (a) =>
      a.platformKeyApi
        .list({ scope: 'member', page: 1, pageSize: 50 })
        .then((page) => page.items.filter((key) => key.status === 'active')),
    enabled: open,
  })

  const selectedKeyId = useMemo(() => {
    if (!open || platformKeys.length === 0) return userSelectedKeyId
    if (userSelectedKeyId && platformKeys.some((key) => key.id === userSelectedKeyId)) {
      return userSelectedKeyId
    }
    const storedId = readStoredPlatformKeyId()
    return platformKeys.find((key) => key.id === storedId)?.id ?? platformKeys[0]?.id ?? ''
  }, [open, platformKeys, userSelectedKeyId])

  const {
    data: bearer = '',
    loading: resolvingKey,
    error: bearerQueryError,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.keys.all, 'simulate-bearer', selectedKeyId] as const,
    queryFn: async (a) => {
      const sk = await resolveBearer(a.devApi, selectedKeyId)
      writeStoredPlatformKeyId(selectedKeyId)
      return sk
    },
    enabled: open && Boolean(selectedKeyId),
  })

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

  const error =
    submitError ??
    (keysQueryError ? '无法加载 Platform Key 列表' : null) ??
    (bearerQueryError instanceof Error ? bearerQueryError.message : null)

  const selectPlatformKey = useCallback((platformKeyId: string) => {
    setUserSelectedKeyId(platformKeyId)
    writeStoredPlatformKeyId(platformKeyId)
    setSubmitError(null)
  }, [])

  const handleSubmit = useCallback(async () => {
    if (!selectedKeyId) {
      setSubmitError('请选择 Platform Key')
      return
    }
    if (inputTokens < 1) {
      setSubmitError('Input tokens 须 ≥ 1')
      return
    }

    let sk = bearer.trim()
    if (!sk) {
      try {
        sk = await resolveBearer(apis.devApi, selectedKeyId)
      } catch (err) {
        setSubmitError(err instanceof Error ? err.message : '无法获取 Platform Key')
        return
      }
    }

    setPhase('calling')
    setSubmitError(null)
    setMatchedCall(null)

    try {
      const baselineIds = await fetchBaselineCallIds(apis.auditApi)
      await postChatCompletions({ bearer: sk, inputTokens, outputTokens })
      toast.message('已调用，等待入账…')
      setPhase('waiting')

      const call = await pollForNewCall(apis.auditApi, baselineIds, inputTokens, outputTokens)
      if (!call) {
        setSubmitError('轮询超时：请检查 Ingest Worker 是否在运行')
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
        setSubmitError(err.body || `Gateway 预检失败 (${err.status})`)
      } else {
        setSubmitError(err instanceof Error ? err.message : '提交失败')
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
