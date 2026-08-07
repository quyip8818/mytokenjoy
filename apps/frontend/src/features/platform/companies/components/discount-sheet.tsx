import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { PlatformCompanyOverview } from '@/api/types'
import type { UseCompanyDiscountsReturn } from '../hooks/use-company-discounts'

interface Props {
  target: PlatformCompanyOverview | null
  discountState: UseCompanyDiscountsReturn
  onClose: () => void
}

function formatDiscount(d: number): string {
  if (d === 1) return '无优惠'
  if (d < 1) return `${Math.round(d * 100)}% (${Math.round((1 - d) * 100)}% off)`
  return `${Math.round(d * 100)}% (加价 ${Math.round((d - 1) * 100)}%)`
}

export function DiscountSheet({ target, discountState, onClose }: Props) {
  const { discounts, loading, submit, submitting } = discountState

  const [modelType, setModelType] = useState('*')
  const [discount, setDiscount] = useState('0.80')
  const [note, setNote] = useState('')

  const handleSubmit = async () => {
    const d = Number(discount)
    if (d <= 0 || !modelType.trim()) return
    await submit({ modelType: modelType.trim(), discount: d, note: note || undefined })
    setNote('')
  }

  return (
    <Sheet open={!!target} onOpenChange={(open) => !open && onClose()}>
      <SheetContent data-testid="discount-sheet" className="sm:max-w-md overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{target?.name} — 模型优惠配置</SheetTitle>
        </SheetHeader>

        <div className="mt-4 space-y-6">
          {/* Current discounts */}
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-2">当前优惠</h4>
            {loading ? (
              <p className="text-sm text-muted-foreground">加载中…</p>
            ) : discounts.length === 0 ? (
              <p className="text-sm text-muted-foreground">暂无优惠配置</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>模型</TableHead>
                    <TableHead>折扣</TableHead>
                    <TableHead>备注</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {discounts.map((d) => (
                    <TableRow key={d.modelType}>
                      <TableCell className="font-mono text-xs">
                        {d.modelType === '*' ? '所有模型' : d.modelType}
                      </TableCell>
                      <TableCell className="tabular-nums">{formatDiscount(d.discount)}</TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {d.note || '-'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>

          {/* Add/update form */}
          <div className="space-y-3 border-t pt-4">
            <h4 className="text-sm font-medium text-muted-foreground">新增 / 更新优惠</h4>

            <label className="block text-sm">
              <span className="text-muted-foreground">模型类型</span>
              <Input
                className="mt-1"
                value={modelType}
                onChange={(e) => setModelType(e.target.value)}
                placeholder="* = 所有模型，或输入具体模型名"
              />
              <span className="text-xs text-muted-foreground mt-0.5 block">
                输入 * 表示通配（未单独配置的模型统一适用）
              </span>
            </label>

            <label className="block text-sm">
              <span className="text-muted-foreground">折扣系数</span>
              <Input
                type="number"
                step="0.01"
                min="0.01"
                className="mt-1"
                value={discount}
                onChange={(e) => setDiscount(e.target.value)}
                placeholder="0.80 = 八折"
              />
              <span className="text-xs text-muted-foreground mt-0.5 block">
                0.8 = 八折, 1.0 = 无优惠, 1.2 = 加价20%
              </span>
            </label>

            <label className="block text-sm">
              <span className="text-muted-foreground">备注</span>
              <Input
                className="mt-1"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="合同优惠"
              />
            </label>

            <Button
              data-testid="discount-submit"
              className="w-full"
              disabled={submitting}
              onClick={handleSubmit}
            >
              {submitting ? '保存中…' : '保存'}
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
