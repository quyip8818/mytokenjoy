import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { SimulateConsumeDialog } from '@/features/dev'

function HeaderDevBackendToolbarContent() {
  const [simulateOpen, setSimulateOpen] = useState(false)

  return (
    <>
      <Button variant="outline" size="sm" type="button" onClick={() => setSimulateOpen(true)}>
        模拟消耗
      </Button>
      <SimulateConsumeDialog open={simulateOpen} onOpenChange={setSimulateOpen} />
    </>
  )
}

export function HeaderDevBackendToolbar() {
  if (!import.meta.env.DEV) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
