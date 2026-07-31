import { forwardRef, type SelectHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

// ponytail: 原生 <select> wrapper，用于简单的 option 列表场景（不需要 Radix Select 的搜索/虚拟化）。
// 升级路径：如果需要搜索/分组/虚拟滚动，换成 radix Select 组件。
export const NativeSelect = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, ...props }, ref) => (
    <select
      ref={ref}
      className={cn(
        'flex h-9 w-full rounded-md border border-input bg-background px-2 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    />
  ),
)
NativeSelect.displayName = 'NativeSelect'
