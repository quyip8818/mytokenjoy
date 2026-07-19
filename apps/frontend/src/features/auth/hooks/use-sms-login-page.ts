import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { authApi, type SmsVerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSession } from '@/features/session'
import { ROUTES } from '@/config/routes'

const COUNTDOWN_SECONDS = 60

export function useSmsLoginPage() {
  const navigate = useNavigate()
  const { refreshSession } = useSession()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [sending, setSending] = useState(false)
  const [verifying, setVerifying] = useState(false)
  const [countdown, setCountdown] = useState(0)

  // Countdown timer for send cooldown.
  const isCountingDown = countdown > 0
  useEffect(() => {
    if (!isCountingDown) return
    const id = setInterval(() => {
      setCountdown((prev) => (prev <= 1 ? 0 : prev - 1))
    }, 1000)
    return () => clearInterval(id)
  }, [isCountingDown])

  const handleSendCode = useCallback(async () => {
    if (!phone.trim() || countdown > 0) return
    setSending(true)
    setError(null)
    try {
      await authApi.smsSend(phone.trim())
      setCountdown(COUNTDOWN_SECONDS)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
        if (err.retryAfter) {
          setCountdown(err.retryAfter)
        }
      } else {
        setError('发送失败，请重试')
      }
    } finally {
      setSending(false)
    }
  }, [phone, countdown])

  const handleVerify = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!phone.trim() || !code.trim()) return
      setVerifying(true)
      setError(null)
      try {
        const result: SmsVerifyResult = await authApi.smsVerify(phone.trim(), code.trim())
        switch (result.action) {
          case 'enter':
            await refreshSession()
            navigate(ROUTES.home, { replace: true })
            break
          case 'select_company':
            navigate('/login/select', { state: { companies: result.companies }, replace: true })
            break
          case 'choose':
            navigate('/onboard', {
              state: { invites: result.invites, mode: 'choose' },
              replace: true,
            })
            break
          case 'onboard':
            navigate('/onboard', { state: { mode: 'onboard' }, replace: true })
            break
        }
      } catch (err) {
        const message = err instanceof ApiError ? err.message : '验证失败'
        setError(message)
      } finally {
        setVerifying(false)
      }
    },
    [phone, code, navigate, refreshSession],
  )

  return {
    phone,
    setPhone,
    code,
    setCode,
    error,
    sending,
    verifying,
    countdown,
    canSend: phone.trim().length >= 11 && countdown === 0 && !sending,
    handleSendCode,
    handleVerify,
  }
}
