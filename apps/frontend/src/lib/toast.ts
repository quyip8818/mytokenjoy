/**
 * ponytail: 统一 toast 控制器。
 * - duration 按级别：error 5s, info/warning 3s, success 2s
 * - error 按 code 匹配 deep link action（来自 ApiError）
 * - 升级路径：sonner 支持 per-type duration 后可简化 duration 逻辑
 */
import { toast as sonnerToast, type ExternalToast } from 'sonner'
import { ApiError } from '@/api/client'

type ToastMsg = string | Parameters<typeof sonnerToast.success>[0]

// ─── Deep link 规则表（按 error code 匹配）───
interface DeepLinkRule {
  code: string
  label: string
  route: string
}

const ERROR_DEEP_LINKS: DeepLinkRule[] = [
  // 预算
  { code: 'BUDGET_RESERVED_POOL_INSUFFICIENT', label: '前往预算管理', route: '/budget' },
  { code: 'BUDGET_DEPT_POOL_INSUFFICIENT', label: '前往预算管理', route: '/budget' },
  { code: 'BUDGET_PROJECT_UNALLOCATED_INSUFFICIENT', label: '前往预算管理', route: '/budget' },
  { code: 'BUDGET_EXCEED_PARENT', label: '前往预算管理', route: '/budget' },
  { code: 'BUDGET_BELOW_ALLOCATED', label: '前往预算管理', route: '/budget' },
  { code: 'BUDGET_DEPT_NOT_SET', label: '前往预算管理', route: '/budget' },
  // Key
  { code: 'KEY_BUDGET_INSUFFICIENT', label: '前往预算管理', route: '/budget' },
  { code: 'KEY_MODEL_DISABLED', label: '前往模型管理', route: '/models' },
  { code: 'KEY_MODEL_NOT_FOUND', label: '前往模型管理', route: '/models' },
  // 试用
  { code: 'TRIAL_MEMBER_LIMIT', label: '升级', route: '/upgrade' },
  { code: 'TRIAL_NO_TOPUP', label: '升级', route: '/upgrade' },
]

function resolveDeepLink(err: unknown): ExternalToast | undefined {
  const code = err instanceof ApiError ? err.code : undefined
  if (!code) return undefined
  const rule = ERROR_DEEP_LINKS.find((r) => r.code === code)
  if (!rule) return undefined
  return {
    action: {
      label: rule.label + ' →',
      onClick: () => (window.location.href = rule.route),
    },
  }
}

// ─── Exported toast API ───
// toast.error 接受 string（手动消息）或 ApiError（自动解析 code + message）
export const toast = {
  success: (msg: ToastMsg, opts?: ExternalToast) =>
    sonnerToast.success(msg, { duration: 2000, ...opts }),

  error: (msgOrError: ToastMsg | ApiError, opts?: ExternalToast) => {
    const message = msgOrError instanceof ApiError ? msgOrError.message : msgOrError
    const link = resolveDeepLink(msgOrError)
    return sonnerToast.error(message, { duration: 5000, ...link, ...opts })
  },

  info: (msg: ToastMsg, opts?: ExternalToast) => sonnerToast.info(msg, { duration: 3000, ...opts }),

  warning: (msg: ToastMsg, opts?: ExternalToast) =>
    sonnerToast.warning(msg, { duration: 3000, ...opts }),

  loading: sonnerToast.loading,
  dismiss: sonnerToast.dismiss,
  promise: sonnerToast.promise,
} as const
