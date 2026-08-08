import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { SimulateConsumeDialog } from '@/features/billing'
import { useSession } from '@/features/session'
import { isTestingAccount } from '@/lib/company'

function HeaderDevBackendToolbarContent() {
  const [simulateOpen, setSimulateOpen] = useState(false)

  return (
    <>
      <Button variant="outline" type="button" onClick={() => setSimulateOpen(true)}>
        模拟消耗
      </Button>
      <SimulateConsumeDialog open={simulateOpen} onOpenChange={setSimulateOpen} />
    </>
  )
}

export function HeaderDevBackendToolbar() {
  const { companyType } = useSession()
  if (!isTestingAccount(companyType)) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
