import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { authApi } from '@/api/auth'
import { ApiError } from '@/api/client'
import { ROUTES } from '@/config/routes'
import { useSession } from '@/features/session'
import { useSmsCountdown } from './use-sms-countdown'

type Step = 'phone' | 'info'

export function useRegisterPage() {
  const navigate = useNavigate()
  const { refreshSession } = useSession()

  // SMS step
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [verifying, setVerifying] = useState(false)

  // Info step
  const [companyName, setCompanyName] = useState('')
  const [password, setPassword] = useState('')
  const [creating, setCreating] = useState(false)

  const [step, setStep] = useState<Step>('phone')
  const [error, setError] = useState<string | null>(null)

  const { sending, countdown, sendError, sendCode } = useSmsCountdown()

  const handleSendCode = useCallback(() => sendCode(phone), [sendCode, phone])

  const handleVerifyAndInit = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!phone.trim() || !code.trim()) return
      setVerifying(true)
      setError(null)
      try {
        const result = await authApi.registerInit(phone.trim(), code.trim())
        if (result.action === 'login') {
          setError('该手机号已注册，请登录')
          return
        }
        // Success — register session cookie set, proceed to info step.
        setStep('info')
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '验证失败')
      } finally {
        setVerifying(false)
      }
    },
    [phone, code],
  )

  const handleCreateCompany = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault()
      if (!companyName.trim() || password.length < 8) return
      setCreating(true)
      setError(null)
      try {
        await authApi.registerCompany(companyName.trim(), password)
        await refreshSession()
        navigate(ROUTES.home, { replace: true })
      } catch (err) {
        setError(err instanceof ApiError ? err.message : '创建失败')
      } finally {
        setCreating(false)
      }
    },
    [companyName, password, navigate, refreshSession],
  )

  return {
    step,
    phone,
    setPhone,
    code,
    setCode,
    companyName,
    setCompanyName,
    password,
    setPassword,
    error: error || sendError,
    sending,
    verifying,
    creating,
    countdown,
    canSend: phone.trim().length >= 11 && countdown === 0 && !sending,
    canSubmitInfo: companyName.trim().length > 0 && password.length >= 8,
    handleSendCode,
    handleVerifyAndInit,
    handleCreateCompany,
  }
}
