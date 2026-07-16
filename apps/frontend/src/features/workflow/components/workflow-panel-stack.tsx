import { useCallback, useEffect, useState } from 'react'
import { useWorkflow } from '../hooks/use-workflow'
import { WorkflowPanelLayer } from './workflow-panel-layer'
import { WorkflowUnsavedDialog } from './workflow-unsaved-dialog'
import type { WorkflowId, WorkflowPayload } from '../types'

export function WorkflowPanelStack() {
  const { stack, dirty, pop, closeAll, push, setDirty } = useWorkflow()
  const [pendingClose, setPendingClose] = useState(false)

  const handleCloseAll = useCallback(() => {
    if (dirty) {
      setPendingClose(true)
      return
    }
    closeAll()
  }, [dirty, closeAll])

  const handlePop = useCallback(() => {
    if (dirty && stack.length === 1) {
      setPendingClose(true)
      return
    }
    pop()
    setDirty(false)
  }, [dirty, stack.length, pop, setDirty])

  const handlePush = useCallback(
    (id: WorkflowId, payload?: WorkflowPayload, title?: string) => {
      push(id, payload, title)
    },
    [push],
  )

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && stack.length > 0) {
        if (stack.length === 1 && dirty) {
          setPendingClose(true)
        } else if (stack.length === 1) {
          closeAll()
        } else {
          pop()
          setDirty(false)
        }
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [stack.length, dirty, closeAll, pop, setDirty])

  if (stack.length === 0) return null

  return (
    <>
      {stack[0]?.layer === 1 && (
        <div className="fixed inset-0 z-30 bg-slate-900/20" onClick={handleCloseAll} aria-hidden />
      )}
      <div className="fixed inset-0 z-30 pointer-events-none">
        <div className="relative h-full w-full">
          {stack.map((entry, index) => (
            <WorkflowPanelLayer
              key={`${entry.id}-${index}`}
              entry={entry}
              index={index}
              total={stack.length}
              onClose={handleCloseAll}
              onPop={handlePop}
              onPush={handlePush}
              onSetDirty={setDirty}
            />
          ))}
        </div>
      </div>

      <WorkflowUnsavedDialog
        open={pendingClose}
        onCancel={() => setPendingClose(false)}
        onConfirm={() => {
          setPendingClose(false)
          closeAll()
        }}
      />
    </>
  )
}
