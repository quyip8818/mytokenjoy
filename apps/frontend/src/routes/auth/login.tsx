import { useNavigate } from 'react-router'
import { useSession } from '@/features/session'
import { ROUTES } from '@/config/routes'
import { AuthPopup, FakeDashboard } from '@/features/auth'

export default function LoginPage() {
  const navigate = useNavigate()
  const { refreshSession } = useSession()

  const handleSuccess = async () => {
    await refreshSession()
    navigate(ROUTES.home, { replace: true })
  }

  return (
    <>
      <FakeDashboard />
      <AuthPopup open={true} defaultMode="login" closable={false} onSuccess={handleSuccess} />
    </>
  )
}
