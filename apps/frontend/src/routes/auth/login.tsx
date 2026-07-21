import { useNavigate } from 'react-router'
import { useSession } from '@/features/session'
import { ROUTES } from '@/config/routes'
import { AuthPopup } from '@/features/auth/components/auth-popup'
import { FakeDashboard } from '@/features/auth/components/fake-dashboard'

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
      <AuthPopup
        open={true}
        defaultMode="login"
        closable={false}
        onSuccess={handleSuccess}
      />
    </>
  )
}
