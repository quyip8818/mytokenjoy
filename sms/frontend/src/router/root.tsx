/* eslint-disable react-refresh/only-export-components */
import { Outlet, createRootRoute } from '@tanstack/react-router'
import { AppProviders } from '@/components/layout/app-providers'

function RootLayout() {
  return (
    <AppProviders>
      <Outlet />
    </AppProviders>
  )
}

export const rootRoute = createRootRoute({
  component: RootLayout,
})
