import { Copy, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '../hooks/use-workflow'

export function KeyRevealWorkflow({ entry, onClose }: WorkflowComponentProps<'key-reveal'>) {
  const { closeAll } = useWorkflow()
  const fullKey = entry.payload.fullKey ?? 'tj-demo-key-not-available'
  const onDone = entry.payload.onDone

  const handleCopy = async () => {
    await navigator.clipboard.writeText(fullKey)
    toast.success('Key 已复制到剪贴板')
  }

  const handleDone = () => {
    onDone?.()
    closeAll()
  }

  return (
    <WorkflowPanelChrome
      title="Key 已生成"
      onClose={onClose}
      footer={<WorkflowPanelFooter primaryLabel="完成" onPrimary={handleDone} />}
    >
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-6">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-50">
          <KeyRound className="h-8 w-8 text-emerald-600" />
        </div>
        <div className="space-y-2">
          <p className="text-sm text-muted-foreground">请立即复制保存，此 Key 仅展示一次</p>
          <WorkflowInfoBox variant="code">
            <span className="flex-1 break-all text-left">{fullKey}</span>
            <Button variant="ghost" size="icon" onClick={handleCopy}>
              <Copy className="h-4 w-4" />
            </Button>
          </WorkflowInfoBox>
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
