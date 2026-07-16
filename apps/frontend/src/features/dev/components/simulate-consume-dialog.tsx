import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { LOCAL_TEST_MODEL } from '../lib/constants'
import { useSimulateConsumeDialog } from '../hooks/use-simulate-consume-dialog'

interface SimulateConsumeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SimulateConsumeDialog({ open, onOpenChange }: SimulateConsumeDialogProps) {
  const dialog = useSimulateConsumeDialog(open, undefined, () => onOpenChange(false))

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>模拟消耗</DialogTitle>
          <DialogDescription>
            调用 <code>{LOCAL_TEST_MODEL}</code> 走 Gateway 预检与转发。HTTP 200
            即表示调用成功；扣费 / 入账由后台 Worker 异步完成。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="local-test-model-key">Platform Key</Label>
            <Select
              value={dialog.selectedKeyId || undefined}
              onValueChange={(value) => void dialog.setSelectedKeyId(value)}
              disabled={dialog.keysLoading || dialog.busy || dialog.platformKeys.length === 0}
            >
              <SelectTrigger id="local-test-model-key" className="w-full">
                <SelectValue
                  placeholder={
                    dialog.keysLoading
                      ? '加载 Key 列表…'
                      : dialog.platformKeys.length === 0
                        ? '无可用 Key'
                        : '选择 Platform Key'
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {dialog.platformKeys.map((key) => (
                  <SelectItem key={key.id} value={key.id}>
                    {dialog.platformKeyOptions[key.id]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {dialog.resolvingKey ? (
              <p className="text-muted-foreground text-xs">正在获取 sk-…</p>
            ) : dialog.platformKeys.length === 0 && !dialog.keysLoading ? (
              <p className="text-muted-foreground text-xs">
                请先在 /keys/platform 创建 active 的 member Key，且白名单含 {LOCAL_TEST_MODEL}
              </p>
            ) : null}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="local-test-model-input">Input tokens</Label>
              <Input
                id="local-test-model-input"
                inputMode="numeric"
                value={dialog.inputTokensText}
                onChange={(e) => dialog.setInputTokensText(e.target.value)}
                disabled={dialog.busy}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="local-test-model-output">Output tokens</Label>
              <Input
                id="local-test-model-output"
                inputMode="numeric"
                value={dialog.outputTokensText}
                onChange={(e) => dialog.setOutputTokensText(e.target.value)}
                disabled={dialog.busy}
              />
            </div>
          </div>

          <p className="text-muted-foreground text-sm">
            预估：<span className="text-foreground font-medium">{dialog.estimatedCost}</span>
          </p>

          {dialog.error ? (
            <pre className="bg-destructive/10 text-destructive max-h-40 overflow-auto rounded-md p-3 text-xs whitespace-pre-wrap break-all">
              {dialog.error}
            </pre>
          ) : null}
        </div>

        <DialogFooter>
          <Button
            type="button"
            onClick={() => void dialog.handleSubmit()}
            disabled={dialog.busy || !dialog.selectedKeyId}
          >
            {dialog.busy ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                提交中…
              </>
            ) : (
              '提交'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
