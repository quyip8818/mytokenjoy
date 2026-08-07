import { Coins, Pencil, Power } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
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
import { CurrencyHistoryDialog } from './currency-history-dialog'
import type { usePlatformCurrenciesPage } from '../hooks/use-platform-currencies-page'

type Props = ReturnType<typeof usePlatformCurrenciesPage>

function fmtQpu(n: number) {
  return n.toLocaleString('zh-CN')
}

function fmtTime(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  return (
    d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) +
    ' ' +
    d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  )
}

export function PlatformCurrenciesPageShell(props: Props) {
  const {
    currencies,
    loading,
    error,
    refresh,
    showCreate,
    createCode,
    setCreateCode,
    createQpu,
    setCreateQpu,
    creating,
    openCreate,
    closeCreate,
    handleCreate,
    editTarget,
    editQpu,
    setEditQpu,
    editing,
    openEdit,
    closeEdit,
    handleEdit,
    handleToggleStatus,
  } = props

  return (
    <PageShell>
      <PageHeader
        testId="page-platform-currencies"
        title="汇率管理"
        icon={Coins}
        actions={
          <div className="flex items-center gap-2">
            <CurrencyHistoryDialog currencies={currencies} />
            <Button onClick={openCreate}>+ 新增币种</Button>
          </div>
        }
      />
      <Card>
        <CardContent className="p-0">
          <DataSection
            loading={loading}
            error={error}
            onRetry={refresh}
            skeletonColumns={6}
            empty={
              currencies.length === 0
                ? { title: '暂无币种', description: '还没有配置任何币种' }
                : null
            }
          >
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>币种代码</TableHead>
                  <TableHead className="text-right">Quota/单位</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>修改人</TableHead>
                  <TableHead>修改时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {currencies.map((c) => (
                  <TableRow key={c.code}>
                    <TableCell className="font-mono font-medium">{c.code}</TableCell>
                    <TableCell className="text-right tabular-nums">
                      {fmtQpu(c.quotaPerUnit)}
                    </TableCell>
                    <TableCell>
                      {c.enabled ? (
                        <Badge variant="default">启用</Badge>
                      ) : (
                        <Badge variant="secondary">禁用</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {c.updatedByName ?? '—'}
                    </TableCell>
                    <TableCell className="text-muted-foreground" title={c.updatedAt}>
                      {fmtTime(c.updatedAt)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="h-8 w-8"
                          title="编辑 Quota"
                          onClick={() => openEdit(c)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="h-8 w-8"
                          title={c.enabled ? '禁用' : '启用'}
                          onClick={() => handleToggleStatus(c)}
                        >
                          <Power className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </DataSection>
          <p className="px-4 pb-3 text-xs text-muted-foreground">
            修改 quota_per_unit 仅影响后续充值，已有充值不受影响。
          </p>
        </CardContent>
      </Card>

      {/* Create dialog */}
      {showCreate && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={closeCreate}
        >
          <div
            className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-base font-semibold">新增币种</h3>
            <label className="mt-4 block text-sm">
              <span className="text-muted-foreground">币种代码（3 位大写字母）</span>
              <input
                type="text"
                className="mt-1 w-full rounded-md border px-3 py-2 text-sm uppercase"
                placeholder="USD"
                maxLength={3}
                value={createCode}
                onChange={(e) => setCreateCode(e.target.value.toUpperCase())}
                autoFocus
              />
            </label>
            <label className="mt-3 block text-sm">
              <span className="text-muted-foreground">Quota/单位</span>
              <input
                type="number"
                className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
                placeholder="500000"
                value={createQpu}
                onChange={(e) => setCreateQpu(e.target.value)}
              />
            </label>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="outline" onClick={closeCreate}>
                取消
              </Button>
              <Button disabled={creating} onClick={handleCreate}>
                {creating ? '创建中…' : '确认创建'}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Edit QPU dialog */}
      {editTarget && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={closeEdit}
        >
          <div
            className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-base font-semibold">编辑 {editTarget.code}</h3>
            <p className="mt-1 text-xs text-muted-foreground">
              修改 quota_per_unit 将影响所有后续充值的额度换算。已有充值（lot）不受影响。
            </p>
            <label className="mt-4 block text-sm">
              <span className="text-muted-foreground">Quota/单位</span>
              <input
                type="number"
                className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
                value={editQpu}
                onChange={(e) => setEditQpu(e.target.value)}
                autoFocus
              />
            </label>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="outline" onClick={closeEdit}>
                取消
              </Button>
              <Button disabled={editing} onClick={handleEdit}>
                {editing ? '更新中…' : '确认修改'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </PageShell>
  )
}
