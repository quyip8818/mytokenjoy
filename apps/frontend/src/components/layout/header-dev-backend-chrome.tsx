import { Link } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
import { Button } from '@/components/ui/button'
import { useSession } from '@/features/session/use-session'

function HeaderDevBackendToolbarContent() {
  const { member } = useSession()

  return (
    <div className="flex items-center gap-2">
      {member && (
        <span className="hidden text-sm text-muted-foreground sm:inline">{member.name}</span>
      )}
      <Button variant="outline" size="sm" asChild>
        <Link to={LOGIN_PATH}>Switch member</Link>
      </Button>
    </div>
  )
}

export function HeaderDevBackendToolbar() {
  if (!import.meta.env.DEV) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
