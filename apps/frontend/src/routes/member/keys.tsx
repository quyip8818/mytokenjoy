import { useState } from 'react'
import { Copy, Check, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { mockPlatformKeys } from '@/mocks/data'

const CURRENT_USER_ID = 'm-1'
const myKeys = mockPlatformKeys.filter((k) => k.memberId === CURRENT_USER_ID)

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <Button
      variant="ghost"
      size="icon"
      className="size-6"
      aria-label="复制"
      onClick={() => { void navigator.clipboard.writeText(text).then(() => { setCopied(true); setTimeout(() => setCopied(false), 1500) }) }}
    >
      {copied ? <Check className="size-3.5 text-emerald-600" /> : <Copy className="size-3.5" />}
    </Button>
  )
}

export default function MemberKeysPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newQuota, setNewQuota] = useState('')

  const handleCreate = () => {
    if (!newName || !newQuota) return
    setNewName(''); setNewQuota(''); setCreateOpen(false)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-sm font-semibold">我的 Key</h1>
        <Button size="sm" className="gap-1.5" onClick={() => setCreateOpen(true)}>
          <Plus className="size-3.5" />
          新建 Key
        </Button>
      </div>

      <div className="rounded-lg border border-border bg-card shadow-xs">
        {myKeys.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-muted-foreground">暂无 Key，点击上方按钮创建</p>
        ) : (
          <div className="divide-y divide-border">
            {myKeys.map((key) => {
              const pct = key.quota > 0 ? Math.round((key.used / key.quota) * 100) : 0
              return (
                <div key={key.id} className="flex items-center gap-4 px-5 py-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">{key.name}</span>
                      <Badge variant="outline" className={cn(
                        'text-xs',
                        key.status === 'active' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-red-50 text-red-700 border-red-200'
                      )}>
                        {key.status === 'active' ? '启用' : '禁用'}
                      </Badge>
                    </div>
                    <div className="mt-1 flex items-center gap-2">
                      <code className="text-xs text-muted-foreground font-mono">{key.keyPrefix}</code>
                      <CopyButton text={key.keyPrefix} />
                    </div>
                    <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                      <span>创建于 {key.createdAt}</span>
                      {key.expiresAt && <span>· 到期 {key.expiresAt}</span>}
                    </div>
                  </div>
                  <div className="w-40 shrink-0">
                    <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                      <span>额度使用</span>
                      <span className="tabular-nums">{pct}%</span>
                    </div>
                    <Progress value={pct} className="h-1.5" />
                    <p className="mt-1 text-xs tabular-nums text-muted-foreground text-right">
                      ¥{key.used.toLocaleString()} / ¥{key.quota.toLocaleString()}
                    </p>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Create Key Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>新建 Key</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-2">
            <div className="grid gap-1.5">
              <Label className="text-xs font-medium">Key 名称</Label>
              <Input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="输入名称" className="h-8 text-sm" />
            </div>
            <div className="grid gap-1.5">
              <Label className="text-xs font-medium">额度（元）</Label>
              <Input type="number" min="1" value={newQuota} onChange={(e) => setNewQuota(e.target.value)} placeholder="输入额度" className="h-8 text-sm tabular-nums" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setCreateOpen(false)}>取消</Button>
            <Button size="sm" onClick={handleCreate}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
