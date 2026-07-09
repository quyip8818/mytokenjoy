import { describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import type { AppApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider, createTestQueryClient } from '@/features/query'
import { AuthSessionProvider, SessionGate, useSession } from '@/features/session'
import { LOGIN_PATH } from '@/config/auth'
import { createMockApis, createMockSession } from '@tests/utils'
import { ApiError } from '@/api/client'

function SessionErrorProbe() {
  const { sessionError } = useSession()
  if (sessionError instanceof ApiError && sessionError.status === 401) {
    return <div data-testid="unauthorized">unauthorized</div>
  }
  return null
}

function renderAuthSession(
  overrides: Partial<AppApis['sessionApi']>,
  options?: { withSessionGate?: boolean },
) {
  const apis = createMockApis({ sessionApi: overrides })
  const content = options?.withSessionGate ? (
    <SessionGate>
      <div data-testid="child">app</div>
    </SessionGate>
  ) : (
    <>
      <SessionErrorProbe />
      <div data-testid="child">app</div>
    </>
  )

  return render(
    <MemoryRouter initialEntries={['/']}>
      <QueryProvider client={createTestQueryClient()}>
        <ApiProvider apis={apis}>
          <Routes>
            <Route path={LOGIN_PATH.slice(1)} element={<div data-testid="login">login</div>} />
            <Route
              path="*"
              element={<AuthSessionProvider apis={apis}>{content}</AuthSessionProvider>}
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
    const replace = vi.fn()
    vi.stubGlobal('location', { ...window.location, replace })

    renderAuthSession(
      {
        getCurrent: vi.fn().mockRejectedValue(new ApiError(401, 'Unauthorized')),
      },
      { withSessionGate: true },
    )

    await waitFor(() => {
      expect(replace).toHaveBeenCalledWith(LOGIN_PATH)
    })
    expect(screen.queryByTestId('child')).not.toBeInTheDocument()
    expect(screen.queryByText('Unauthorized')).not.toBeInTheDocument()

    vi.unstubAllGlobals()
  })

  it('exposes 401 sessionError without SessionGate', async () => {
    renderAuthSession({
      getCurrent: vi.fn().mockRejectedValue(new ApiError(401, 'Unauthorized')),
    })

    await waitFor(() => {
      expect(screen.getByTestId('unauthorized')).toBeInTheDocument()
    })
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('renders children when getCurrent returns incomplete payload', async () => {
    renderAuthSession({
      getCurrent: vi.fn().mockResolvedValue({ invalid: true }),
    })

    await waitFor(() => {
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })
})
