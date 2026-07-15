import { useState } from 'react'
import { Link } from 'react-router'
import { LOGIN_PATH } from '@/config/auth'
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
      <Button variant="outline" size="sm" asChild>
        <Link to={LOGIN_PATH}>Switch member</Link>
      </Button>
    </>
  )
}

export function HeaderDevBackendToolbar() {
  if (!import.meta.env.DEV) {
    return null
  }

  return <HeaderDevBackendToolbarContent />
}
