import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** 计算距离到期日的天数，负数=已过期 */
export function daysUntil(endDate?: string): number | null {
  if (!endDate) return null
  const end = new Date(endDate).getTime()
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return Math.ceil((end - now.getTime()) / (24 * 3600 * 1000))
}

/** 金额格式化（¥ 两位小数） */
export function formatAmount(amount?: number | null): string {
  if (amount === undefined || amount === null) return '-'
  return Number(amount).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}
