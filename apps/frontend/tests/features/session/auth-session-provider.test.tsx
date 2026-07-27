import { describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { useEffect, useRef } from 'react'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import type { AppApis } from '@/api/app-apis'
import { ApiProvider } from '@/api/context'
import { QueryProvider, createTestQueryClient } from '@/features/query'
import { AuthSessionProvider, useSession } from '@/features/session'
import { ApiError } from '@/api/client'
import { LOGIN_PATH } from '@/config/auth'
import { createMockApis, createMockSession } from '@tests/utils'

function SessionErrorProbe() {
  const { sessionError } = useSession()
  if (sessionError instanceof ApiError && sessionError.status === 401) {
    return <div data-testid="unauthorized">unauthorized</div>
  }
  return null
}

/**
 * Minimal inline equivalent of the old SessionGate for testing.
 * Redirects to login on 401 via window.location.replace (same as auth-layout).
 */
function TestAuthGate({ children }: { children: React.ReactNode }) {
  const { sessionError, loading } = useSession()
  const hasRedirected = useRef(false)
  const isUnauthorized = sessionError instanceof ApiError && sessionError.status === 401

  useEffect(() => {
    if (isUnauthorized && !hasRedirected.current) {
      hasRedirected.current = true
      window.location.replace(LOGIN_PATH)
    }
  }, [isUnauthorized])

  if (loading) return null
  if (isUnauthorized) return null
  return <>{children}</>
}

function renderAuthSession(
  overrides: Partial<AppApis['sessionApi']>,
  options?: { withAuthGate?: boolean },
) {
  const apis = createMockApis({ sessionApi: overrides })
  const content = options?.withAuthGate ? (
    <TestAuthGate>
      <div data-testid="child">app</div>
    </TestAuthGate>
  ) : (
    <>
      <SessionErrorProbe />
      <div data-testid="child">app</div>
    </>
  )

  const rootRoute = createRootRoute({
    component: () => (
      <QueryProvider client={createTestQueryClient()}>
        <ApiProvider apis={apis}>
          <AuthSessionProvider apis={apis}>{content}</AuthSessionProvider>
        </ApiProvider>
      </QueryProvider>
    ),
  })
  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: LOGIN_PATH,
    component: () => <div data-testid="login">login</div>,
  })
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
  })
  const catchAll = createRoute({
    getParentRoute: () => rootRoute,
    path: '$',
  })

  const routeTree = rootRoute.addChildren([loginRoute, indexRoute, catchAll])
  const history = createMemoryHistory({ initialEntries: ['/'] })
  const router = createRouter({ routeTree, history })

  return render(<RouterProvider router={router} />)
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
      { withAuthGate: true },
    )

    await waitFor(() => {
      expect(replace).toHaveBeenCalledWith(LOGIN_PATH)
    })
    expect(screen.queryByTestId('child')).not.toBeInTheDocument()
    expect(screen.queryByText('Unauthorized')).not.toBeInTheDocument()

    vi.unstubAllGlobals()
  })

  it('exposes 401 sessionError without auth gate', async () => {
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
