import { useEffect, useState } from 'react'
import { Monitor } from 'lucide-react'

const DESKTOP_MIN_WIDTH_PX = 1024

export function DesktopOnlyHint() {
  const [show, setShow] = useState(false)

  useEffect(() => {
    const mq = window.matchMedia(`(max-width: ${DESKTOP_MIN_WIDTH_PX - 1}px)`)
    const update = () => setShow(mq.matches)
    update()
    mq.addEventListener('change', update)
    return () => mq.removeEventListener('change', update)
  }, [])

  if (!show) return null

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-background/95 p-6">
      <div className="max-w-sm text-center space-y-4">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-50">
          <Monitor className="h-6 w-6 text-blue-600" />
        </div>
        <p className="text-base font-medium">请使用桌面浏览器</p>
        <p className="text-sm text-muted-foreground">
          TokenJoy Demo 针对桌面端优化，建议在宽度 1024px 以上的浏览器中演示。
        </p>
      </div>
    </div>
  )
}
