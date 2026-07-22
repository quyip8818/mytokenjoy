import { useCallback, useMemo, useRef, useState } from 'react'
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
import { compressImage, ImageValidationError } from '@/lib/compress-image'

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
  avataaars: '卡通人物',
  lorelei: '手绘',
  micah: '扁平',
  adventurer: '冒险者',
  notionists: '简笔',
  'big-smile': '笑脸',
  'open-peeps': '涂鸦',
  'pixel-art': '像素',
}

const GRID_COUNT = 12

export function AvatarPicker({ value, onChange, trigger, size = 48 }: AvatarPickerProps) {
  const [open, setOpen] = useState(false)
  const [style, setStyle] = useState<AvatarStyle>('adventurer')
  const [seeds, setSeeds] = useState<string[]>(() => generateSeeds())
  const [selected, setSelected] = useState<string>(value)
  const [uploading, setUploading] = useState(false)
  const [uploadError, setUploadError] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  function generateSeeds() {
    return Array.from({ length: GRID_COUNT }, () => randomSeed())
  }

  const handleOpen = useCallback(() => {
    setSelected(value)
    setUploadError('')
    setOpen(true)
  }, [value])

  const handleConfirm = useCallback(() => {
    onChange(selected)
    setOpen(false)
  }, [selected, onChange])

  const handleShuffle = useCallback(() => {
    setSeeds(generateSeeds())
  }, [])

  const handleUpload = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    // Reset input so same file can be re-selected
    e.target.value = ''

    setUploadError('')
    setUploading(true)
    try {
      const dataUri = await compressImage(file)
      setSelected(dataUri)
    } catch (err) {
      if (err instanceof ImageValidationError) {
        setUploadError(err.message)
      } else {
        setUploadError('图片处理失败，请重试')
      }
    } finally {
      setUploading(false)
    }
  }, [])

  const previews = useMemo(
    () =>
      seeds.map((seed) => ({
        seed,
        value: dicebearValue(style, seed),
        dataUri: renderAvatar(dicebearValue(style, seed)),
      })),
    [style, seeds],
  )

  const displayUri = renderAvatar(value)

  return (
    <>
      {trigger ? (
        <span onClick={handleOpen} className="cursor-pointer">
          {trigger}
        </span>
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
        <DialogContent className="sm:max-w-[460px]">
          <DialogTitle className="text-base font-semibold">选择头像</DialogTitle>

          {/* Style tabs + upload button */}
          <div className="flex flex-wrap items-center gap-1.5 mt-2">
            {AVATAR_STYLE_NAMES.map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => {
                  setStyle(s)
                  setSeeds(generateSeeds())
                }}
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
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              className="px-2.5 py-1 text-xs rounded-md transition-colors bg-muted text-muted-foreground hover:bg-accent border border-dashed border-border"
            >
              {uploading ? '处理中…' : '上传图片'}
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/webp"
              className="hidden"
              onChange={handleUpload}
            />
          </div>

          <p className="text-xs text-muted-foreground mt-1">
            支持 PNG、JPEG、WebP 格式，不超过 5MB
          </p>
          {uploadError && <p className="text-xs text-destructive mt-0.5">{uploadError}</p>}

          {/* Grid */}
          <div className="grid grid-cols-4 gap-2.5 mt-4">
            {/* Show uploaded/custom preview as first cell if selected is a data URI */}
            {selected.startsWith('data:') && (
              <button
                type="button"
                className="aspect-square rounded-lg border-2 overflow-hidden transition-all p-1 border-primary ring-2 ring-primary/20"
              >
                <img
                  src={selected}
                  alt="uploaded"
                  className="w-full h-full rounded object-cover"
                />
              </button>
            )}
            {previews.map((p) => (
              <button
                key={p.seed}
                type="button"
                onClick={() => setSelected(p.value)}
                className={cn(
                  'aspect-square rounded-lg border-2 overflow-hidden transition-all p-2',
                  selected === p.value
                    ? 'border-primary ring-2 ring-primary/20'
                    : 'border-border hover:border-primary/50',
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
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  onChange('')
                  setOpen(false)
                }}
              >
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
