import { useState } from 'react'

export type StructureConfirmState = {
  open: boolean
  title: string
  desc: string
  variant: 'primary' | 'danger'
  onConfirm: () => void
}

export const INITIAL_STRUCTURE_CONFIRM_STATE: StructureConfirmState = {
  open: false,
  title: '',
  desc: '',
  variant: 'primary',
  onConfirm: () => {},
}

export function useStructureConfirmState() {
  const [confirmState, setConfirmState] = useState<StructureConfirmState>(
    INITIAL_STRUCTURE_CONFIRM_STATE,
  )

  const closeConfirm = () => {
    setConfirmState((state) => ({ ...state, open: false }))
  }

  return { confirmState, setConfirmState, closeConfirm }
}
