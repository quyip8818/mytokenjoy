import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { authApi, type SmsVerifyResult } from '@/api/auth'
import { ApiError } from '@/api/client'
import { useSession } from '@/features/session'
import { ROUTES } from '@/config/routes'
import { useSmsCountdown } from './use-sms-countdown'

export function useSmsLoginPage() {
  const navigate = useNavigate()
  const { refreshSession } = useSession()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [showRegisterHint, setShowRegisterHint] = useState(false)

  const { sending, countdown, sendError, sendCode } = useSmsCountdown()

  const handleSendCode = useCallback(() => sendCode(phone), [sendCode, phone])

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
          case 'not_found':
            setError('该手机号尚未注册，请先注册')
            setShowRegisterHint(true)
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
    error: error || sendError,
    sending,
    verifying,
    countdown,
    showRegisterHint,
    canSend: phone.trim().length >= 11 && countdown === 0 && !sending,
    handleSendCode,
    handleVerify,
  }
}
