import { request } from './client'

/**
 * Dev-only API (`/api/dev/*`). Backend registers these routes only when
 * DEPLOY_ENV=local (see config.AllowsDevHTTPRoutes). Do not call from production
 * builds or add staging/production fallbacks.
 */
export const devApi = {
  getPlatformKeyBearer: (platformKeyId: string) => {
    if (!import.meta.env.DEV) {
      throw new Error('devApi is only available in Vite development builds')
    }
    return request<{ bearer: string }>(`/dev/platform-keys/${platformKeyId}/bearer`)
  },
}
