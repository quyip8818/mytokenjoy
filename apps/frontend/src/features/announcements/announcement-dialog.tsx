import { ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import type { AnnouncementConfig } from './types'
import { useAnnouncement } from './use-announcement'

/** Generic announcement dialog — renders whatever config useAnnouncement returns. */
export function AnnouncementDialog() {
  const { current, open, close } = useAnnouncement()

  if (!current) return null

  return (
    <Dialog open={open} onOpenChange={(v) => !v && close()}>
      <DialogContent className="sm:max-w-md" showCloseButton={false}>
        <AnnouncementBody config={current} onClose={close} />
      </DialogContent>
    </Dialog>
  )
}

function AnnouncementBody({
  config,
  onClose,
}: {
  config: AnnouncementConfig
  onClose: () => void
}) {
  const Icon = config.icon

  return (
    <>
      {/* Header gradient */}
      <div
        className={`-mx-4 -mt-4 rounded-t-lg bg-gradient-to-br ${config.gradient} px-6 py-8 text-white`}
      >
        <div className="flex items-center gap-3">
          <div className="flex size-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm">
            <Icon className="size-6" />
          </div>
          <div>
            <h2 className="text-lg font-semibold">{config.title}</h2>
            <p className="text-sm text-white/80">{config.subtitle}</p>
          </div>
        </div>
      </div>

      <DialogHeader className="pt-2">
        <DialogTitle className="sr-only">{config.title}</DialogTitle>
        <DialogDescription className="sr-only">{config.subtitle}</DialogDescription>
      </DialogHeader>

      {/* Feature cards */}
      <div className="space-y-3">
        {config.features.map((f) => {
          const FIcon = f.icon
          return (
            <div key={f.title} className={`flex items-start gap-3 rounded-lg border p-3 ${f.cardClass}`}>
              <div className={`flex size-8 shrink-0 items-center justify-center rounded-lg ${f.iconClass}`}>
                <FIcon className="size-4" />
              </div>
              <div>
                <p className="text-sm font-medium text-foreground">{f.title}</p>
                <p className="text-xs text-muted-foreground">{f.description}</p>
              </div>
            </div>
          )
        })}
      </div>

      {/* CTA */}
      <div className="flex justify-end pt-1">
        <Button variant="brand" className="gap-1.5" onClick={onClose}>
          {config.ctaLabel}
          <ArrowRight className="size-3.5" />
        </Button>
      </div>
    </>
  )
}
