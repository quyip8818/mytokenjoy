import { useCallback, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'
import {
  AVATAR_STYLE_NAMES,
  type AvatarStyle,
  dicebearValue,
  randomSeed,
  renderAvatar,
} from '@/lib/avatar'

interface AvatarPickerProps {
  /** Current avatar value (dicebear:... or data:image/...) */
  value: string
  /** Called when user confirms selection */
  onChange: (avatar: string) => void
  /** Optional trigger element. If not provided, renders a clickable avatar circle. */
  trigger?: React.ReactNode
  /** Size of the display avatar in px */
  size?: number
}

const STYLE_LABELS: Record<AvatarStyle, string> = {
  adventurer: '冒险者',
  notionists: '简笔',
  bottts: '机器人',
  shapes: '几何',
  lorelei: '手绘',
  'fun-emoji': '表情',
}

const GRID_COUNT = 9

export function AvatarPicker({ value, onChange, trigger, size = 48 }: AvatarPickerProps) {
  const [open, setOpen] = useState(false)
  const [style, setStyle] = useState<AvatarStyle>('adventurer')
  const [seeds, setSeeds] = useState<string[]>(() => generateSeeds())
  const [selected, setSelected] = useState<string>(value)

  function generateSeeds() {
    return Array.from({ length: GRID_COUNT }, () => randomSeed())
  }

  const handleOpen = useCallback(() => {
    setSelected(value)
    setOpen(true)
  }, [value])

  const handleConfirm = useCallback(() => {
    onChange(selected)
    setOpen(false)
  }, [selected, onChange])

  const handleShuffle = useCallback(() => {
    setSeeds(generateSeeds())
  }, [])

  const previews = useMemo(
    () => seeds.map((seed) => ({ seed, value: dicebearValue(style, seed), dataUri: renderAvatar(dicebearValue(style, seed)) })),
    [style, seeds],
  )

  const displayUri = renderAvatar(value)

  return (
    <>
      {trigger ? (
        <span onClick={handleOpen} className="cursor-pointer">{trigger}</span>
      ) : (
        <button
          type="button"
          onClick={handleOpen}
          className="rounded-full border-2 border-dashed border-border hover:border-primary transition-colors flex items-center justify-center overflow-hidden bg-muted"
          style={{ width: size, height: size }}
          aria-label="选择头像"
        >
          {displayUri ? (
            <img src={displayUri} alt="avatar" className="w-full h-full object-cover" />
          ) : (
            <span className="text-xs text-muted-foreground">头像</span>
          )}
        </button>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogTitle className="text-base font-semibold">选择头像</DialogTitle>

          {/* Style tabs */}
          <div className="flex flex-wrap gap-1.5 mt-2">
            {AVATAR_STYLE_NAMES.map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => { setStyle(s); setSeeds(generateSeeds()) }}
                className={cn(
                  'px-2.5 py-1 text-xs rounded-md transition-colors',
                  style === s
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground hover:bg-accent',
                )}
              >
                {STYLE_LABELS[s]}
              </button>
            ))}
          </div>

          {/* Grid */}
          <div className="grid grid-cols-3 gap-3 mt-4">
            {previews.map((p) => (
              <button
                key={p.seed}
                type="button"
                onClick={() => setSelected(p.value)}
                className={cn(
                  'aspect-square rounded-lg border-2 overflow-hidden transition-all p-2',
                  selected === p.value ? 'border-primary ring-2 ring-primary/20' : 'border-border hover:border-primary/50',
                )}
              >
                <img src={p.dataUri} alt={p.seed} className="w-full h-full" />
              </button>
            ))}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between mt-4">
            <Button type="button" variant="ghost" size="sm" onClick={handleShuffle}>
              换一批
            </Button>
            <div className="flex gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => { onChange(''); setOpen(false) }}>
                清除
              </Button>
              <Button type="button" size="sm" onClick={handleConfirm}>
                确认
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
