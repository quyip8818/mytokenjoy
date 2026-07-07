import { Link } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
import { Button } from '@/components/ui/button'

function HeaderDevBackendToolbarContent() {
  return (
    <Button variant="outline" size="sm" asChild>
      <Link to={LOGIN_PATH}>Switch member</Link>
    </Button>
  )
}

export function HeaderDevBackendToolbar() {
  if (!import.meta.env.DEV) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
