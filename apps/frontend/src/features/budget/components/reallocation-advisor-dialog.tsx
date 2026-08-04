import { useCallback, useMemo, useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { AlertTriangle, Check, X } from 'lucide-react'
import { formatMoney } from '@/lib/quota-display'
import { cn } from '@/lib/utils'

export interface AllocationItem {
  key: string
  label: string
  originalValue: number
}

interface ReallocationAdvisorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Original total budget */
  oldBudget: number
  /** New (lower) total budget */
  newBudget: number
  /** Breakdown of current allocations */
  items: AllocationItem[]
  /** Called with the final allocation values when user applies */
  onApply: (allocations: Record<string, number>) => void
  /** Called when user chooses to manually adjust (only sets total, doesn't touch sub-items) */
  onManual: () => void
}

/**
 * ponytail: reallocation strategy is "eat reserved pool first, then proportional cut".
 * The "reserved pool" is identified by key='reservedPool' in items.
 */
function computeSuggestion(items: AllocationItem[], newBudget: number): Record<string, number> {
  const totalAllocated = items.reduce((sum, item) => sum + item.originalValue, 0)
  let gap = totalAllocated - newBudget
  if (gap <= 0) {
    // No reallocation needed
    return Object.fromEntries(items.map((item) => [item.key, item.originalValue]))
  }

  const result: Record<string, number> = {}

  // Step 1: Eat reserved pool first
  const reservedItem = items.find((item) => item.key === 'reservedPool')
  if (reservedItem && reservedItem.originalValue > 0) {
    const reservedCut = Math.min(gap, reservedItem.originalValue)
    gap -= reservedCut
    result['reservedPool'] = reservedItem.originalValue - reservedCut
  }

  // Step 2: Proportional cut on remaining items
  const otherItems = items.filter((item) => item.key !== 'reservedPool')
  const otherTotal = otherItems.reduce((sum, item) => sum + item.originalValue, 0)

  if (gap > 0 && otherTotal > 0) {
    const ratio = Math.max(0, (otherTotal - gap) / otherTotal)
    for (const item of otherItems) {
      result[item.key] = Math.max(0, Math.round(item.originalValue * ratio * 100) / 100)
    }
  } else {
    for (const item of otherItems) {
      result[item.key] = item.originalValue
    }
  }

  return result
}

export function ReallocationAdvisorDialog({
  open,
  onOpenChange,
  oldBudget,
  newBudget,
  items,
  onApply,
  onManual,
}: ReallocationAdvisorDialogProps) {
  const suggestion = useMemo(() => computeSuggestion(items, newBudget), [items, newBudget])

  const [drafts, setDrafts] = useState<Record<string, string>>({})

  // Reset drafts when dialog opens
  const handleOpenChange = useCallback(
    (value: boolean) => {
      if (value) {
        // Pre-fill with suggestions
        const initial: Record<string, string> = {}
        for (const item of items) {
          initial[item.key] = String(suggestion[item.key] ?? item.originalValue)
        }
        setDrafts(initial)
      }
      onOpenChange(value)
    },
    [items, suggestion, onOpenChange],
  )

  const gap = items.reduce((sum, item) => sum + item.originalValue, 0) - newBudget

  const currentTotal = useMemo(() => {
    return items.reduce((sum, item) => {
      const raw = drafts[item.key] ?? String(suggestion[item.key] ?? item.originalValue)
      const val = parseFloat(raw)
      return sum + (Number.isNaN(val) ? 0 : val)
    }, 0)
  }, [items, drafts, suggestion])

  const isValid =
    currentTotal <= newBudget &&
    !items.some((item) => {
      const raw = drafts[item.key] ?? String(suggestion[item.key] ?? item.originalValue)
      const val = parseFloat(raw)
      return Number.isNaN(val) || val < 0
    })

  const handleApply = () => {
    const allocations: Record<string, number> = {}
    for (const item of items) {
      const raw = drafts[item.key] ?? String(suggestion[item.key] ?? item.originalValue)
      allocations[item.key] = parseFloat(raw) || 0
    }
    onApply(allocations)
    onOpenChange(false)
  }

  const handleManual = () => {
    onManual()
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="size-5 text-amber-500" />
            预算缩减需要重新分配
          </DialogTitle>
        </DialogHeader>

        {/* Summary line */}
        <div className="flex items-center gap-3 text-sm">
          <span className="text-muted-foreground">
            部门总预算 {formatMoney(oldBudget)} →{' '}
            <span className="font-semibold text-foreground">{formatMoney(newBudget)}</span>
          </span>
          <span className="text-red-500 font-medium">缺口: {formatMoney(gap)}</span>
        </div>

        {/* Strategy hint */}
        <p className="text-xs text-muted-foreground">
          策略：优先消耗预留池，剩余按原比例等额分摊至各子项
        </p>

        {/* Allocation table */}
        <div className="mt-2 divide-y divide-border rounded-lg border border-border">
          {/* Header */}
          <div className="grid grid-cols-[1fr_80px_80px_80px_100px] gap-2 px-3 py-2 text-xs font-medium text-muted-foreground">
            <span>项目</span>
            <span className="text-right">原值</span>
            <span className="text-right">缩减</span>
            <span className="text-right">建议值</span>
            <span className="text-right">手动调整</span>
          </div>

          {/* Rows */}
          {items.map((item) => {
            const suggested = suggestion[item.key] ?? item.originalValue
            const cut = item.originalValue - suggested
            return (
              <div
                key={item.key}
                className="grid grid-cols-[1fr_80px_80px_80px_100px] items-center gap-2 px-3 py-2"
              >
                <span className="text-sm font-medium text-foreground truncate">{item.label}</span>
                <span className="text-right text-sm tabular-nums text-muted-foreground">
                  {formatMoney(item.originalValue)}
                </span>
                <span
                  className={cn(
                    'text-right text-sm tabular-nums',
                    cut > 0 ? 'text-red-500' : 'text-muted-foreground',
                  )}
                >
                  {cut > 0 ? `-${formatMoney(cut)}` : '—'}
                </span>
                <span className="text-right text-sm tabular-nums font-medium">
                  {formatMoney(suggested)}
                </span>
                <Input
                  type="number"
                  min={0}
                  step="any"
                  value={drafts[item.key] ?? String(suggested)}
                  onChange={(e) => setDrafts((prev) => ({ ...prev, [item.key]: e.target.value }))}
                  className="h-7 w-full text-right tabular-nums text-sm"
                />
              </div>
            )
          })}

          {/* Total row */}
          <div className="grid grid-cols-[1fr_80px_80px_80px_100px] items-center gap-2 px-3 py-2 bg-muted/30">
            <span className="text-sm font-semibold">合计</span>
            <span className="text-right text-sm tabular-nums font-semibold">
              {formatMoney(items.reduce((s, i) => s + i.originalValue, 0))}
            </span>
            <span />
            <span />
            <span
              className={cn(
                'flex items-center justify-end gap-1 text-sm tabular-nums font-semibold',
                isValid ? 'text-foreground' : 'text-red-500',
              )}
            >
              {formatMoney(currentTotal)}
              {isValid ? (
                <Check className="size-3.5 text-emerald-500" />
              ) : (
                <X className="size-3.5 text-red-500" />
              )}
            </span>
          </div>
        </div>

        {/* Validation error */}
        {!isValid && (
          <p className="text-xs text-red-500">
            合计超出新预算 {formatMoney(currentTotal - newBudget)}，请调整
          </p>
        )}

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button variant="outline" onClick={handleManual}>
            我手动分配
          </Button>
          <Button onClick={handleApply} disabled={!isValid}>
            应用建议方案
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
