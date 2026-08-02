import type { ReactNode } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'

/** Minimum model shape required by ModelTable. */
export interface ModelTableRow {
  modelId: string
  type: string
  name: string
  provider: string
  inputPrice: number
  outputPrice: number
  deprecated: boolean
}

interface ModelTableProps<T extends ModelTableRow> {
  models: T[]
  /** Extra columns between price and actions (optional). */
  extraColumns?: { header: string; render: (model: T) => ReactNode }[]
  /** Actions column (optional). If omitted, no actions column rendered. */
  renderActions?: (model: T) => ReactNode
  /** Optional row class name resolver. */
  rowClass?: (model: T) => string | undefined
}

function formatPrice(price: number) {
  if (price <= 0) return <span className="text-muted-foreground">—</span>
  return (
    <>
      <span className="font-mono text-xs tabular-nums text-foreground/80">¥{price}</span>
      <span className="ml-0.5 text-[10px] text-muted-foreground">/M</span>
    </>
  )
}

export function ModelTable<T extends ModelTableRow>({
  models,
  extraColumns,
  renderActions,
  rowClass,
}: ModelTableProps<T>) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent border-border/60">
          <TableHead className="text-xs font-medium text-muted-foreground">模型</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">Provider</TableHead>
          <TableHead className="text-right text-xs font-medium text-muted-foreground">
            输入价格
          </TableHead>
          <TableHead className="text-right text-xs font-medium text-muted-foreground">
            输出价格
          </TableHead>
          {extraColumns?.map((col) => (
            <TableHead key={col.header} className="text-xs font-medium text-muted-foreground">
              {col.header}
            </TableHead>
          ))}
          {renderActions && (
            <TableHead className="w-[140px] text-right text-xs font-medium text-muted-foreground">
              操作
            </TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {models.map((model) => (
          <TableRow
            key={model.modelId}
            className={cn(
              'group transition-colors',
              model.deprecated && 'opacity-45',
              rowClass?.(model),
            )}
          >
            <TableCell>
              <div className="flex flex-col gap-0.5">
                <span className="text-sm font-medium leading-tight text-foreground">
                  {model.name}
                </span>
                <span className="font-mono text-xs text-muted-foreground/80">{model.type}</span>
              </div>
            </TableCell>
            <TableCell className="text-muted-foreground text-sm">{model.provider}</TableCell>
            <TableCell className="text-right">{formatPrice(model.inputPrice)}</TableCell>
            <TableCell className="text-right">{formatPrice(model.outputPrice)}</TableCell>
            {extraColumns?.map((col) => (
              <TableCell key={col.header}>{col.render(model)}</TableCell>
            ))}
            {renderActions && (
              <TableCell className="text-right">{renderActions(model)}</TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
