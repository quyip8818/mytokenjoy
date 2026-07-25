import { LogOut } from 'lucide-react'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'

export function Header() {
  const { user, logout } = useSession()
  const { authApi } = useApis()

  const handleLogout = async () => {
    try {
      await authApi.logout()
    } catch {
      /* ignore */
    }
    logout()
  }

  return (
    <header className="flex h-14 items-center justify-end border-b bg-white px-6">
      <div className="flex items-center gap-3">
        <span className="text-sm text-muted-foreground">{user?.realName}</span>
        <button
          onClick={handleLogout}
          className="flex items-center gap-1 rounded-md px-2 py-1 text-sm text-muted-foreground hover:bg-muted"
        >
          <LogOut className="h-4 w-4" />
          退出
        </button>
      </div>
    </header>
  )
}
