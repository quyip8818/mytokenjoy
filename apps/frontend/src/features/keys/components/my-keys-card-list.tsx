import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import type { PlatformKey } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  return (
    <Button
      variant="ghost"
      size="icon"
      className="size-6"
      aria-label="复制"
      onClick={() => {
        void navigator.clipboard.writeText(text).then(() => {
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        })
      }}
    >
      {copied ? <Check className="size-3.5 text-emerald-600" /> : <Copy className="size-3.5" />}
    </Button>
  )
}

interface MyKeysCardListProps {
  keys: PlatformKey[]
}

export function MyKeysCardList({ keys }: MyKeysCardListProps) {
  if (keys.length === 0) {
    return (
      <p className="px-5 py-8 text-center text-sm text-muted-foreground">
        暂无 Key，点击上方按钮创建
      </p>
    )
  }

  return (
    <div className="divide-y divide-border">
      {keys.map((key) => {
        const pct = key.budget > 0 ? Math.round((key.consumed / key.budget) * 100) : 0
        return (
          <div key={key.id} className="flex items-center gap-4 px-5 py-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className="truncate text-sm font-medium">{key.name}</span>
                <Badge
                  variant="outline"
                  className={cn(
                    'text-xs',
                    key.status === 'active'
                      ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                      : 'border-red-200 bg-red-50 text-red-700',
                  )}
                >
                  {key.status === 'active' ? '启用' : '禁用'}
                </Badge>
              </div>
              <div className="mt-1 flex items-center gap-2">
                <code className="font-mono text-xs text-muted-foreground">{key.keyPrefix}</code>
                <CopyButton text={key.keyPrefix} />
              </div>
              <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                <span>创建于 {key.createdAt}</span>
                {key.expiresAt && <span>· 到期 {key.expiresAt}</span>}
              </div>
            </div>
            <div className="w-40 shrink-0">
              <div className="mb-1 flex items-center justify-between text-xs text-muted-foreground">
                <span>额度使用</span>
                <span className="tabular-nums">{pct}%</span>
              </div>
              <Progress value={pct} className="h-1.5" />
              <p className="mt-1 text-right text-xs text-muted-foreground tabular-nums">
                ¥{key.consumed.toLocaleString()} / ¥{key.budget.toLocaleString()}
              </p>
            </div>
          </div>
        )
      })}
    </div>
  )
}
