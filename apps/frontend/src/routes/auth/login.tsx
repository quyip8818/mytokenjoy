import { useNavigate } from 'react-router'
import { Link } from 'react-router'
import { ROUTES } from '@/config/routes'
import { DEV_SWITCHABLE_MEMBERS } from '@/config/dev-members'
import { EmptyState } from '@/components/ui/empty-state'
import { Button } from '@/components/ui/button'
import { setSessionMemberCookie } from '@/lib/session-cookie'

function DevLoginPanel() {
  const navigate = useNavigate()

  return (
    <div className="flex w-full max-w-md flex-col gap-6">
      <div className="space-y-2 text-center">
        <h1 className="text-lg font-semibold">Dev backend sign-in</h1>
        <p className="text-sm text-muted-foreground">
          Select a member to set the session cookie. API calls use same-origin `/api` proxied to the
          Go backend.
        </p>
      </div>
      <ul className="flex flex-col gap-2">
        {DEV_SWITCHABLE_MEMBERS.map((member) => (
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
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      {import.meta.env.DEV ? (
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
