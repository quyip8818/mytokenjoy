import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import {
  WORKFLOW_ANIMATION_MS,
  WORKFLOW_LAYER_MAX_WIDTH,
  WORKFLOW_LAYER_WIDTH,
  WORKFLOW_PEEK_WIDTH_PX,
} from '../constants'
import type { WorkflowId, WorkflowLayer, WorkflowPayload, WorkflowStackEntry } from '../types'
import { getWorkflowDefinition } from '../workflow-definitions'

interface WorkflowPanelLayerProps {
  entry: WorkflowStackEntry
  index: number
  total: number
  onClose: () => void
  onPop: () => void
  onPush: (id: WorkflowId, payload?: WorkflowPayload, title?: string) => void
  onSetDirty: (dirty: boolean) => void
}

const LAYER_STYLES: Record<WorkflowLayer, { width: string; maxWidth: number }> = {
  1: { width: WORKFLOW_LAYER_WIDTH.layer1, maxWidth: WORKFLOW_LAYER_MAX_WIDTH.layer1 },
  2: { width: WORKFLOW_LAYER_WIDTH.layer2, maxWidth: WORKFLOW_LAYER_MAX_WIDTH.layer2 },
  3: { width: WORKFLOW_LAYER_WIDTH.layer3, maxWidth: WORKFLOW_LAYER_MAX_WIDTH.layer3 },
}

export function WorkflowPanelLayer({
  entry,
  index,
  total,
  onClose,
  onPop,
  onPush,
  onSetDirty,
}: WorkflowPanelLayerProps) {
  const [visible, setVisible] = useState(false)
  const isTop = index === total - 1
  const depthFromTop = total - 1 - index
  const styles = LAYER_STYLES[entry.layer]
  const def = getWorkflowDefinition(entry.id)
  const Component = def.component

  useEffect(() => {
    const timer = requestAnimationFrame(() => setVisible(true))
    return () => cancelAnimationFrame(timer)
  }, [])

  const handleClose = () => {
    if (isTop) {
      if (index === 0) onClose()
      else onPop()
    }
  }

  const handleBack = () => {
    onPop()
  }

  return (
    <div
      className={cn(
        'absolute top-0 bottom-0 flex flex-col border-l border-border/80 bg-white transition-transform ease-out',
        depthFromTop > 0 && 'shadow-lg',
        index === 0 && 'shadow-[var(--shadow-sidebar)]',
      )}
      style={{
        right: 0,
        width: styles.width,
        maxWidth: styles.maxWidth,
        zIndex: 40 + index,
        transform: visible ? 'translateX(0)' : 'translateX(100%)',
        transitionDuration: `${WORKFLOW_ANIMATION_MS}ms`,
        marginRight: depthFromTop > 0 ? depthFromTop * WORKFLOW_PEEK_WIDTH_PX : 0,
        pointerEvents: isTop ? 'auto' : 'none',
      }}
    >
      {depthFromTop > 0 && (
        <div
          className="absolute left-0 top-0 bottom-0 w-3 bg-slate-50 border-r border-border/60"
          aria-hidden
        />
      )}
      <div className="flex h-full flex-col pl-0">
        <Component
          entry={entry}
          onClose={handleClose}
          onPop={handleBack}
          onPush={onPush}
          onSetDirty={onSetDirty}
        />
      </div>
    </div>
  )
}
