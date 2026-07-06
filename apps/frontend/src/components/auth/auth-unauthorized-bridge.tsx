import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { setUnauthorizedHandler } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'

export function AuthUnauthorizedBridge() {
  const navigate = useNavigate()

  useEffect(() => {
    setUnauthorizedHandler(() => {
      navigate(LOGIN_PATH, { replace: true })
    })
    return () => setUnauthorizedHandler(null)
  }, [navigate])

  return null
}
