import { useNavigate } from 'react-router'
import { CheckCircle2, Circle, Map } from 'lucide-react'
import { DEMO_GUIDE_STEPS } from './constants'
import { useDemoGuide } from '@/features/demo'
import { useDemoRole } from '@/features/demo/roles/use-demo-role'
import { DEMO_ROLE_PROFILES } from '@/features/demo/roles/constants'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export function DemoGuidePanel() {
  const navigate = useNavigate()
  const { role, setRole } = useDemoRole()
  const {
    open,
    setOpen,
    highlightCtaId,
    setHighlightCtaId,
    markComplete,
    resetProgress,
    completed,
  } = useDemoGuide()

  const completedCount = DEMO_GUIDE_STEPS.filter((s) => completed.has(s.id)).length

  const handleGoToStep = (step: (typeof DEMO_GUIDE_STEPS)[number]) => {
    if (step.role && step.role !== role) {
      setRole(step.role)
    }
    setOpen(false)
    navigate(step.path)
    if (step.ctaId) {
      const ctaId = step.ctaId
      window.setTimeout(() => {
        setHighlightCtaId(ctaId)
        const el = document.getElementById(ctaId)
        el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }, 300)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Map className="h-5 w-5 text-indigo-600" />
            演示引导
          </DialogTitle>
          <DialogDescription>
            按步骤完成管理员初始化与成员闭环演示（{completedCount}/{DEMO_GUIDE_STEPS.length}）
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2 py-2">
          {DEMO_GUIDE_STEPS.map((step, index) => {
            const done = completed.has(step.id)
            const isActive = step.ctaId && highlightCtaId === step.ctaId
            return (
              <div
                key={step.id}
                className={cn(
                  'flex items-start gap-3 rounded-lg border border-border/60 p-3 transition-colors',
                  isActive && 'border-indigo-300 bg-indigo-50/50',
                  done && 'bg-slate-50/50',
                )}
              >
                <div className="mt-0.5 shrink-0">
                  {done ? (
                    <CheckCircle2 className="h-5 w-5 text-emerald-600" />
                  ) : (
                    <Circle className="h-5 w-5 text-muted-foreground/40" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">
                    {index + 1}. {step.title}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">{step.description}</p>
                  {step.role && (
                    <p className="text-xs text-indigo-600 mt-1">
                      视角：{DEMO_ROLE_PROFILES[step.role].label}
                    </p>
                  )}
                </div>
                <div className="flex flex-col gap-1 shrink-0">
                  <Button size="sm" variant="outline" onClick={() => handleGoToStep(step)}>
                    前往
                  </Button>
                  {!done && (
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-xs h-7"
                      onClick={() => markComplete(step.id)}
                    >
                      完成
                    </Button>
                  )}
                </div>
              </div>
            )
          })}
        </div>

        <div className="flex justify-end gap-2 pt-2 border-t">
          <Button variant="outline" size="sm" onClick={resetProgress}>
            重置进度
          </Button>
          <Button size="sm" onClick={() => setOpen(false)}>
            关闭
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
