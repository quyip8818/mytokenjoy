import { describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider, createTestQueryClient } from '@/features/query'
import { AuthSessionProvider } from '@/features/session'
import { LOGIN_PATH } from '@/config/auth'
import { createMockApis, createMockSession } from '@tests/utils'
import { ApiError } from '@/api/client'

function renderAuthSession(overrides: Partial<AppApis['sessionApi']>) {
  const apis = createMockApis({ sessionApi: overrides })

  return render(
    <MemoryRouter initialEntries={['/']}>
      <QueryProvider client={createTestQueryClient()}>
        <ApiProvider apis={apis}>
          <Routes>
            <Route path={LOGIN_PATH.slice(1)} element={<div data-testid="login">login</div>} />
            <Route
              path="*"
              element={
                <AuthSessionProvider apis={apis}>
                  <div data-testid="child">app</div>
                </AuthSessionProvider>
              }
            />
          </Routes>
        </ApiProvider>
      </QueryProvider>
    </MemoryRouter>,
  )
}

describe('AuthSessionProvider', () => {
  it('renders children when getCurrent succeeds', async () => {
    const session = createMockSession()
    renderAuthSession({
      getCurrent: vi.fn().mockResolvedValue(session),
    })

    await waitFor(() => {
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })

  it('redirects to login when getCurrent returns 401', async () => {
    renderAuthSession({
      getCurrent: vi.fn().mockRejectedValue(new ApiError(401, 'Unauthorized')),
    })

    await waitFor(() => {
      expect(screen.getByTestId('login')).toBeInTheDocument()
    })
    expect(screen.queryByText('Unauthorized')).not.toBeInTheDocument()
  })

  it('shows error state when getCurrent returns invalid payload', async () => {
    renderAuthSession({
      getCurrent: vi.fn().mockResolvedValue({ invalid: true }),
    })

    await waitFor(() => {
      expect(screen.getByText('Invalid session response')).toBeInTheDocument()
    })
  })
})
