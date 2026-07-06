import type { PlatformKey } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { AlertTriangle } from 'lucide-react'

export function KeyRotateConfirmWorkflow({
  entry,
  onPop,
  onClose,
  onPush,
}: WorkflowComponentProps<'key-rotate-confirm'>) {
  const key = entry.payload.key as PlatformKey
  const onRotate = entry.payload.onRotate as
    | ((key: PlatformKey) => Promise<{ fullKey?: string; keyPrefix: string }>)
    | undefined
  const onDone = entry.payload.onDone as (() => void) | undefined

  const handleConfirm = async () => {
    if (!onRotate) return
    const rotated = await onRotate(key)
    onPush('key-reveal', {
      fullKey: rotated.fullKey ?? `${rotated.keyPrefix}rotated`,
      onDone,
    })
  }

  return (
    <WorkflowPanelChrome
      title="重新生成 Key"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel="确认重新生成"
          onPrimary={handleConfirm}
        />
      }
    >
      <WorkflowFormLayout
        variant="narrow"
        className="flex flex-col items-center text-center py-8 mx-auto space-y-4"
      >
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-amber-50">
          <AlertTriangle className="h-6 w-6 text-amber-600" />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">确定重新生成「{key.name}」的 Key？</p>
          <p className="text-sm text-muted-foreground">
            旧 Key 将立即失效，使用旧 Key 的应用将无法继续调用。
          </p>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
