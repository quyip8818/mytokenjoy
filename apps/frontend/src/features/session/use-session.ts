import { useContext } from 'react'
import { SessionReactContext } from './context'

export function useSession() {
  const session = useContext(SessionReactContext)
  if (!session) {
    throw new Error('useSession must be used within DemoSessionProvider or AuthSessionProvider')
  }
  return session
}
