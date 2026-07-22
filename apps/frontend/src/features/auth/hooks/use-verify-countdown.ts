import { useCallback, useEffect, useState } from 'react'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'

const COUNTDOWN_SECONDS = 60

/**
 * Generalized countdown hook for verification code sending (phone or email).
 */
export function useVerifyCountdown() {
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
    async (params: { phone?: string; email?: string }) => {
      if (countdown > 0) return
      const hasValue = (params.phone && params.phone.trim()) || (params.email && params.email.trim())
      if (!hasValue) return
      setSending(true)
      setSendError(null)
      try {
        await authApi.sendCode(params)
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
