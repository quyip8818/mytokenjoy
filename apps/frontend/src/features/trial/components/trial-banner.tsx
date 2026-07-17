import { Info } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function TrialBanner() {
  return (
    <div className="flex items-center justify-between border-b border-indigo-100 bg-indigo-50/80 px-6 py-2 text-sm text-indigo-700">
      <div className="flex items-center gap-2">
        <Info className="h-4 w-4 shrink-0" />
        <span>试用环境 · 使用模拟资金体验，升级后接入真实模型</span>
      </div>
      <Button
        variant="outline"
        size="sm"
        className="h-7 border-indigo-200 text-xs text-indigo-700 hover:bg-indigo-100"
      >
        联系升级
      </Button>
    </div>
  )
}
