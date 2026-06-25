import { Map } from 'lucide-react'
import { DEMO_SWITCHABLE_MEMBERS } from '@/features/demo/roles/constants'
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
  const { memberId, displayName, initials, roles, setMemberId } = useDemoRole()
  const { setOpen } = useDemoGuide()

  return (
    <>
      <div className="flex shrink-0 items-center gap-2">
        <Button variant="outline" size="sm" className="h-8 gap-1.5" onClick={() => setOpen(true)}>
          <Map className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">演示引导</span>
        </Button>
        <div className="hidden h-5 w-px bg-border/60 sm:block" aria-hidden />
        <Select
          value={memberId}
          onValueChange={(v) => {
            if (v) void setMemberId(v)
          }}
        >
          <SelectTrigger className="h-8 w-[11rem] text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {DEMO_SWITCHABLE_MEMBERS.map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className="flex items-center gap-2 rounded-full border border-border/60 py-1 pr-3 pl-1 transition-colors hover:border-blue-200 hover:bg-blue-50/50">
          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-sky-500 text-[10px] font-semibold text-white shadow-sm">
            {initials}
          </div>
          <div className="hidden min-w-0 md:block">
            <span className="block truncate text-sm font-medium text-foreground/80">{displayName}</span>
            <span className="block max-w-[10rem] truncate text-[10px] text-muted-foreground">
              {roles.join(' · ')}
            </span>
          </div>
        </div>
      </div>
      <DemoGuidePanel />
    </>
  )
}
