import { User } from 'lucide-react'
import { cn } from '@/lib/utils'
import { renderAvatar } from '@/lib/avatar'

interface UserAvatarProps {
  /** Stored avatar value: dicebear:..., data:image/..., or empty */
  avatar?: string
  /** Fallback text (first char shown when no avatar) */
  fallback?: string
  /** Size in px (default 24) */
  size?: number
  className?: string
}

/**
 * Shared avatar display component.
 * Handles dicebear format, data URI, and fallback to initial or icon.
 */
export function UserAvatar({ avatar, fallback, size = 24, className }: UserAvatarProps) {
  const src = avatar ? renderAvatar(avatar) : ''

  if (src) {
    return (
      <img
        src={src}
        alt=""
        className={cn('rounded-md object-cover', className)}
        style={{ width: size, height: size }}
      />
    )
  }

  const initial = fallback?.charAt(0)
  return (
    <div
      className={cn('flex items-center justify-center rounded-md bg-muted', className)}
      style={{ width: size, height: size }}
    >
      {initial ? (
        <span className="text-[10px] font-medium text-muted-foreground">{initial}</span>
      ) : (
        <User className="h-3.5 w-3.5 text-muted-foreground" />
      )}
    </div>
  )
}
