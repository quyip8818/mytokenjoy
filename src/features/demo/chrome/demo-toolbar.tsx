import { Map } from 'lucide-react'
import { DEMO_ROLES } from '@/features/demo/roles/constants'
import { useDemoRole } from '@/features/demo/roles/use-demo-role'
import { useDemoGuide } from '@/features/demo/guide/use-demo-guide'
import { DemoGuidePanel } from '@/features/demo/guide/demo-guide-panel'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export function DemoToolbar() {
  const { role, displayName, initials, setRole } = useDemoRole()
  const { setOpen } = useDemoGuide()

  return (
    <>
      <Button variant="outline" size="sm" className="h-9 gap-1.5" onClick={() => setOpen(true)}>
        <Map className="h-4 w-4" />
        演示引导
      </Button>
      <Select value={role} onValueChange={(v) => setRole(v as typeof role)}>
        <SelectTrigger className="w-28 h-9">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {DEMO_ROLES.map((r) => (
            <SelectItem key={r.id} value={r.id}>
              {r.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <div className="flex items-center gap-2.5 rounded-full border border-border/60 px-3 py-1.5 transition-colors hover:border-indigo-200 hover:bg-indigo-50/50">
        <div className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-violet-500 text-[11px] font-semibold text-white shadow-sm">
          {initials}
        </div>
        <span className="text-sm font-medium text-foreground/80">{displayName}</span>
      </div>
      <DemoGuidePanel />
    </>
  )
}
