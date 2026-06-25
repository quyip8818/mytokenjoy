import { lazy, Suspense, type ReactNode } from 'react'
import { USE_MOCKS } from '@/config/app'

const LazyDemoShell = lazy(() =>
  import('@/features/demo/demo-shell').then((module) => ({ default: module.DemoShell })),
)

interface LazyDemoShellBoundaryProps {
  children: ReactNode
}

export function LazyDemoShellBoundary({ children }: LazyDemoShellBoundaryProps) {
  if (!USE_MOCKS) {
    return children
  }

  return (
    <Suspense fallback={null}>
      <LazyDemoShell>{children}</LazyDemoShell>
    </Suspense>
  )
}
