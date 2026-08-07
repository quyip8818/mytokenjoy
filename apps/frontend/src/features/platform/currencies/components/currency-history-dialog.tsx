import { useState } from 'react'
import { History } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { PlatformCurrency } from '@/api/types'
import { platformCurrenciesKeys } from '../query-keys'

function fmtTime(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function fmtQpu(n: number) {
  return n.toLocaleString('zh-CN')
}

interface Props {
  currencies: PlatformCurrency[]
}

export function CurrencyHistoryDialog({ currencies }: Props) {
  const [open, setOpen] = useState(false)
  const [selectedCode, setSelectedCode] = useState<string>(currencies[0]?.code ?? '')

  const { data: history = [], loading } = useInjectedQuery({
    queryKey: platformCurrenciesKeys.history(selectedCode),
    queryFn: (apis) => apis.platformApi.listCurrencyHistory(selectedCode),
    enabled: open && !!selectedCode,
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <History className="mr-1.5 h-4 w-4" />
          历史
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>汇率变更历史</DialogTitle>
        </DialogHeader>

        <div className="mb-4">
          <Select value={selectedCode} onValueChange={setSelectedCode}>
            <SelectTrigger className="w-32">
              <SelectValue placeholder="选择币种" />
            </SelectTrigger>
            <SelectContent>
              {currencies.map((c) => (
                <SelectItem key={c.code} value={c.code}>
                  {c.code}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {loading ? (
          <p className="py-8 text-center text-sm text-muted-foreground">加载中…</p>
        ) : history.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">暂无记录</p>
        ) : (
          <div className="max-h-96 overflow-y-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead className="text-right">Quota/单位</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>修改人</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {history.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell className="text-muted-foreground">
                      {fmtTime(row.updatedAt)}
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      {fmtQpu(row.quotaPerUnit)}
                    </TableCell>
                    <TableCell>
                      {row.enabled ? (
                        <Badge variant="default">启用</Badge>
                      ) : (
                        <Badge variant="secondary">禁用</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {row.updatedByName ?? '—'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
