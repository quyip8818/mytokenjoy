/* eslint-disable react-refresh/only-export-components */
import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import { AppErrorBoundary } from '@/components/layout/app-error-boundary'
import { AppProviders } from '@/components/layout/app-providers'
import type { RouterContext } from './context'

function RootLayout() {
  return (
    <AppProviders>
      <AppErrorBoundary>
        <Outlet />
      </AppErrorBoundary>
    </AppProviders>
  )
}

export const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
})
