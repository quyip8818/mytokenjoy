import { useNavigate } from 'react-router'
import { Link } from 'react-router'
import { USE_MOCKS } from '@/config/app'
import { ROUTES } from '@/config/routes'
import { EmptyState } from '@/components/ui/empty-state'
import { Button } from '@/components/ui/button'
import { DEMO_SWITCHABLE_MEMBERS } from '@/features/demo/roles/constants'
import { setSessionMemberCookie } from '@/lib/session-cookie'

function DevLoginPanel() {
  const navigate = useNavigate()

  return (
    <div className="flex w-full max-w-md flex-col gap-6">
      <div className="space-y-2 text-center">
        <h1 className="text-lg font-semibold">Dev backend sign-in</h1>
        <p className="text-sm text-muted-foreground">
          Select a member to set the session cookie and connect to the Go API via Vite proxy.
        </p>
      </div>
      <ul className="flex flex-col gap-2">
        {DEMO_SWITCHABLE_MEMBERS.map((member) => (
          <li key={member.id}>
            <Button
              type="button"
              variant="outline"
              className="h-auto w-full justify-start px-4 py-3 text-left"
              onClick={() => {
                setSessionMemberCookie(member.id)
                navigate(ROUTES.home, { replace: true })
              }}
            >
              <span className="flex flex-col gap-0.5">
                <span className="font-medium">{member.label}</span>
                <span className="text-xs font-normal text-muted-foreground">
                  {member.roleSummary}
                </span>
              </span>
            </Button>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default function LoginPage() {
  const isDevBackend = import.meta.env.DEV && !USE_MOCKS

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      {isDevBackend ? (
        <DevLoginPanel />
      ) : (
        <>
          <EmptyState
            title="Sign in required"
            description="Production authentication is provided by the backend. Configure cookie or JWT login on the API gateway, then return to the admin console."
            className="max-w-md"
          />
          <Link to={ROUTES.home} className="text-sm font-medium text-primary hover:underline">
            Back to home
          </Link>
        </>
      )}
    </div>
  )
}
