import { Building2, MoreHorizontal } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { PlatformCompanyOverview } from '@/api/types'
import type { usePlatformCompaniesPage } from '../hooks/use-platform-companies-page'
import { useWorkflow } from '@/features/workflow'

type Props = ReturnType<typeof usePlatformCompaniesPage>

function fmt(n: number) {
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function StatusBadge({ status }: { status: string }) {
  return status === 'active' ? (
    <Badge variant="default">活跃</Badge>
  ) : (
    <Badge variant="secondary">停用</Badge>
  )
}

function RechargeDialog({
  target,
  amount,
  setAmount,
  loading,
  onConfirm,
  onClose,
}: {
  target: PlatformCompanyOverview
  amount: string
  setAmount: (v: string) => void
  loading: boolean
  onConfirm: () => void
  onClose: () => void
}) {
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>充值</DialogTitle>
          <DialogDescription>{target.name}</DialogDescription>
        </DialogHeader>
        <label className="block text-sm">
          <span className="text-muted-foreground">金额 (元)</span>
          <input
            type="number"
            className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            placeholder="0.00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            autoFocus
          />
        </label>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button disabled={loading} onClick={onConfirm}>
            {loading ? '处理中…' : '确认充值'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function GiftDialog({
  target,
  amount,
  setAmount,
  loading,
  onConfirm,
  onClose,
}: {
  target: PlatformCompanyOverview
  amount: string
  setAmount: (v: string) => void
  loading: boolean
  onConfirm: () => void
  onClose: () => void
}) {
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>赠送额度</DialogTitle>
          <DialogDescription>{target.name}</DialogDescription>
        </DialogHeader>
        <label className="block text-sm">
          <span className="text-muted-foreground">赠送金额 (元)</span>
          <input
            type="number"
            className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            placeholder="0.00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            autoFocus
          />
        </label>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button disabled={loading} onClick={onConfirm}>
            {loading ? '处理中…' : '确认赠送'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// Sort by balance ascending (low balance first).
function sortedByBalance(companies: PlatformCompanyOverview[]) {
  return [...companies].sort((a, b) => a.wallet.balance - b.wallet.balance)
}

export function PlatformCompaniesPageShell(props: Props) {
  const {
    companies,
    loading,
    error,
    refresh,
    rechargeTarget,
    rechargeAmount,
    setRechargeAmount,
    recharging,
    openRecharge,
    closeRecharge,
    handleRecharge,
    giftTarget,
    giftAmount,
    setGiftAmount,
    gifting,
    openGift,
    closeGift,
    handleGift,
    handleToggleStatus,
  } = props

  const { open } = useWorkflow()

  const sorted = sortedByBalance(companies)

  return (
    <PageShell>
      <PageHeader testId="page-platform-companies" title="企业管理" icon={Building2} />
      <Card>
        <CardContent className="p-0">
          <DataSection
            loading={loading}
            error={error}
            onRetry={refresh}
            skeletonColumns={10}
            empty={
              sorted.length === 0 ? { title: '暂无企业', description: '还没有创建任何企业' } : null
            }
          >
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>公司名称</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">余额</TableHead>
                  <TableHead className="text-right">赠送余额</TableHead>
                  <TableHead className="text-right">累计充值</TableHead>
                  <TableHead className="text-right">本月消耗</TableHead>
                  <TableHead className="text-right">成员数</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sorted.map((co) => (
                  <TableRow key={co.id}>
                    <TableCell className="font-medium">{co.name}</TableCell>
                    <TableCell>{co.type}</TableCell>
                    <TableCell>
                      <StatusBadge status={co.status} />
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      ¥{fmt(co.wallet.balance)}
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      ¥{fmt(co.wallet.giftBalance)}
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      ¥{fmt(co.wallet.totalTopup)}
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      ¥{fmt(co.monthlySpend)}
                    </TableCell>
                    <TableCell className="text-right">{co.memberCount}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button variant="outline" onClick={() => openRecharge(co)}>
                          充值
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button size="icon" variant="ghost" className="h-8 w-8">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => openGift(co)}>赠送</DropdownMenuItem>
                            <DropdownMenuItem
                              data-testid={`platform-companies-discount-${co.id}`}
                              onClick={() =>
                                open('discount-config', {
                                  companyId: co.id,
                                  companyName: co.name,
                                  onSuccess: refresh,
                                })
                              }
                            >
                              优惠
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() =>
                                open('lot-audit', {
                                  companyId: co.id,
                                  companyName: co.name,
                                  readonly: false,
                                  onSuccess: refresh,
                                })
                              }
                            >
                              审计
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleToggleStatus(co)}>
                              {co.status === 'active' ? '停用' : '启用'}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </DataSection>
        </CardContent>
      </Card>

      {rechargeTarget && (
        <RechargeDialog
          target={rechargeTarget}
          amount={rechargeAmount}
          setAmount={setRechargeAmount}
          loading={recharging}
          onConfirm={handleRecharge}
          onClose={closeRecharge}
        />
      )}
      {giftTarget && (
        <GiftDialog
          target={giftTarget}
          amount={giftAmount}
          setAmount={setGiftAmount}
          loading={gifting}
          onConfirm={handleGift}
          onClose={closeGift}
        />
      )}
    </PageShell>
  )
}
