import { Link } from 'react-router'
import { ROUTES } from '@/config/routes'
import { EmptyState } from '@/components/ui/empty-state'

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <EmptyState
        title="Sign in required"
        description="Production authentication is provided by the backend. Configure cookie or JWT login on the API gateway, then return to the admin console."
        className="max-w-md"
      />
      <Link to={ROUTES.home} className="text-sm font-medium text-primary hover:underline">
        Back to home
      </Link>
    </div>
  )
}
