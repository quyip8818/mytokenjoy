import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { apiEvents } from '@/api/api-events'
import { LOGIN_PATH } from '@/config/auth'

export function AuthUnauthorizedBridge() {
  const navigate = useNavigate()

  useEffect(() => {
    return apiEvents.on('unauthorized', () => {
      navigate({ to: LOGIN_PATH, replace: true })
    })
  }, [navigate])

  return null
}
