import { screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permission-keys'
import { renderWithProviders } from '@tests/utils'

describe('PermissionGate', () => {
  it('renders children when user has required permission', async () => {
    renderWithProviders(
      <PermissionGate permission={PERMISSION.ORG_STRUCTURE}>
        <span>allowed</span>
      </PermissionGate>,
      { permissions: [PERMISSION.ORG_STRUCTURE, PERMISSION.ORG_MEMBERS] },
    )

    expect(await screen.findByText('allowed')).toBeInTheDocument()
  })

  it('renders fallback when user lacks required permission', async () => {
    renderWithProviders(
      <PermissionGate permission={PERMISSION.ORG_STRUCTURE} fallback={<span>denied</span>}>
        <span>allowed</span>
      </PermissionGate>,
      { permissions: [PERMISSION.SELF_KEYS] },
    )

    await waitFor(() => {
      expect(screen.queryByText('allowed')).not.toBeInTheDocument()
    })
    expect(screen.getByText('denied')).toBeInTheDocument()
  })

  it('hides children in write mode for read-only sessions', async () => {
    renderWithProviders(
      <PermissionGate write fallback={<span>read-only</span>}>
        <span>write-action</span>
      </PermissionGate>,
      {
        permissions: [PERMISSION.AUDIT_READ, PERMISSION.DASHBOARD_COST],
        readOnly: true,
      },
    )

    await waitFor(() => {
      expect(screen.queryByText('write-action')).not.toBeInTheDocument()
    })
    expect(screen.getByText('read-only')).toBeInTheDocument()
  })
})
