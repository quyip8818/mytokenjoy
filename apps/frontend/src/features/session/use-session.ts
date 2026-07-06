import { useContext } from 'react'
import { SessionReactContext } from './context'

export function useSession() {
  const session = useContext(SessionReactContext)
  if (!session) {
    throw new Error('useSession must be used within AuthSessionProvider')
  }
  return session
}
