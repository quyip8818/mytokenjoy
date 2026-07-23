import { useCallback, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import type { PlatformKey } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { toast } from 'sonner'
import {
  DEFAULT_INPUT_TOKENS,
  DEFAULT_OUTPUT_TOKENS,
  PLATFORM_KEY_ID_SESSION_KEY,
  formatEstimatedConsume,
} from '../lib/constants'
import { GatewayClientError, postChatCompletions } from '../lib/simulate-consume'

function readStoredPlatformKeyId(): string {
  return sessionStorage.getItem(PLATFORM_KEY_ID_SESSION_KEY) ?? ''
}

function writeStoredPlatformKeyId(platformKeyId: string) {
  if (platformKeyId) sessionStorage.setItem(PLATFORM_KEY_ID_SESSION_KEY, platformKeyId)
  else sessionStorage.removeItem(PLATFORM_KEY_ID_SESSION_KEY)
}

function parseTokenCount(value: string, fallback: number, min: number): number {
  const parsed = Number.parseInt(value, 10)
  if (!Number.isFinite(parsed) || parsed < min) return fallback
  return parsed
}

function formatPlatformKeyLabel(key: PlatformKey): string {
  const prefix = key.keyPrefix ? ` · ${key.keyPrefix}` : ''
  return `${key.name}${prefix}`
}

async function resolveBearer(
  devApiClient: AppApis['platformKeyApi'],
  platformKeyId: string,
): Promise<string> {
  const { bearer } = await devApiClient.simulateBearer(platformKeyId)
  if (!bearer) {
    throw new Error('未返回 sk-，请检查 Key 是否已同步到 NewAPI')
  }
  return bearer
}

export function useSimulateConsumeDialog(
  open: boolean,
  injectedApis?: AppApis,
  onSuccess?: () => void,
) {
  const apis = useInjectedApis(injectedApis)
  const queryClient = useQueryClient()

  const [userSelectedKeyId, setUserSelectedKeyId] = useState(readStoredPlatformKeyId)
  const [inputTokensText, setInputTokensText] = useState(String(DEFAULT_INPUT_TOKENS))
  const [outputTokensText, setOutputTokensText] = useState(String(DEFAULT_OUTPUT_TOKENS))
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  // When dialog is closed, treat submitError as null (no need for an effect).
  const visibleSubmitError = open ? submitError : null

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
      const sk = await resolveBearer(a.platformKeyApi, selectedKeyId)
      writeStoredPlatformKeyId(selectedKeyId)
      return sk
    },
    enabled: open && Boolean(selectedKeyId),
  })

  const inputTokens = useMemo(
    () => parseTokenCount(inputTokensText, DEFAULT_INPUT_TOKENS, 0),
    [inputTokensText],
  )
  const outputTokens = useMemo(
    () => parseTokenCount(outputTokensText, DEFAULT_OUTPUT_TOKENS, 0),
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
    visibleSubmitError ??
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
        sk = await resolveBearer(apis.platformKeyApi, selectedKeyId)
      } catch (err) {
        setSubmitError(err instanceof Error ? err.message : '无法获取 Platform Key')
        return
      }
    }

    setSubmitting(true)
    setSubmitError(null)

    try {
      await postChatCompletions({ bearer: sk, inputTokens, outputTokens })
      toast.success('Gateway 已受理，入账由 Worker 异步完成')
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.audit.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.wallet.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.budget.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all }),
      ])
      onSuccess?.()
    } catch (err) {
      if (err instanceof GatewayClientError) {
        setSubmitError(err.body || `Gateway 预检失败 (${err.status})`)
      } else {
        setSubmitError(err instanceof Error ? err.message : '提交失败')
      }
    } finally {
      setSubmitting(false)
    }
  }, [apis.platformKeyApi, bearer, inputTokens, onSuccess, outputTokens, queryClient, selectedKeyId])

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
    busy: submitting || resolvingKey,
    error,
    handleSubmit,
  }
}
