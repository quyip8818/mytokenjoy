import { useCallback, useEffect, useState } from 'react'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'

const COUNTDOWN_SECONDS = 60

/**
 * Shared hook for SMS send + countdown logic.
 * Used by both login and register pages.
 */
export function useSmsCountdown() {
  const [sending, setSending] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [sendError, setSendError] = useState<string | null>(null)

  const isCountingDown = countdown > 0
  useEffect(() => {
    if (!isCountingDown) return
    const id = setInterval(() => {
      setCountdown((prev) => (prev <= 1 ? 0 : prev - 1))
    }, 1000)
    return () => clearInterval(id)
  }, [isCountingDown])

  const sendCode = useCallback(
    async (phone: string) => {
      if (!phone.trim() || countdown > 0) return
      setSending(true)
      setSendError(null)
      try {
        await authApi.smsSend(phone.trim())
        setCountdown(COUNTDOWN_SECONDS)
      } catch (err) {
        if (err instanceof ApiError) {
          setSendError(err.message)
          if (err.retryAfter) setCountdown(err.retryAfter)
        } else {
          setSendError('发送失败，请重试')
        }
      } finally {
        setSending(false)
      }
    },
    [countdown],
  )

  return { sending, countdown, sendError, sendCode }
}
