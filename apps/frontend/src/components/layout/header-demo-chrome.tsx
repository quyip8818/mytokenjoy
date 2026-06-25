import { lazy, Suspense } from 'react'
import { USE_MOCKS } from '@/config/app'

const LazyDemoBadge = lazy(() =>
  import('@/features/demo/chrome/demo-banner').then((module) => ({ default: module.DemoBadge })),
)

const LazyDemoToolbar = lazy(() =>
  import('@/features/demo/chrome/demo-toolbar').then((module) => ({ default: module.DemoToolbar })),
)

export function HeaderDemoBadge() {
  if (!USE_MOCKS) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <LazyDemoBadge />
    </Suspense>
  )
}

export function HeaderDemoToolbar() {
  if (!USE_MOCKS) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <LazyDemoToolbar />
    </Suspense>
  )
}
