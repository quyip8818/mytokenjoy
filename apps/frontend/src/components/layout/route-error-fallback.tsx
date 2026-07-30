import { useRouter } from '@tanstack/react-router'
import { captureException } from '@/config/monitoring'
import { ErrorState } from '@/components/ui/error-state'

// ponytail: 路由级错误边界。每个 route 独立捕获渲染异常，不白屏邻居页面。
// 升级路径：按 route 定制错误 UI（传 errorComponent per route）。
export function RouteErrorFallback({ error }: { error: unknown }) {
  const router = useRouter()
  const message = error instanceof Error ? error.message : '发生了未知错误'

  if (import.meta.env.DEV) {
    console.error('RouteErrorFallback caught:', error)
  }
  if (error instanceof Error) {
    captureException(error)
  }

  return (
    <div className="flex min-h-[12rem] items-center justify-center p-8">
      <ErrorState
        title="页面出错"
        message={message}
        onRetry={() => void router.invalidate()}
        retryLabel="重试"
      />
    </div>
  )
}
