import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { SimulateConsumeDialog } from '@/features/dev'
import { useSession } from '@/features/session'

const SIMULATE_ALLOWED_TYPES = ['demo', 'trial', 'testing'] as const

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
  const { companyType } = useSession()
  if (!(SIMULATE_ALLOWED_TYPES as readonly string[]).includes(companyType)) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
