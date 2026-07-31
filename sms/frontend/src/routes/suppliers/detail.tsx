import { useState } from 'react'
import { useParams, Link } from '@tanstack/react-router'
import { ArrowLeft, Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useInjectedQuery, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'
import { SUPPLIER_STATUS, CONTRACT_STATUS, ORDER_STATUS, MODEL_STATUS } from '@/config/enums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { ConfirmActionDialog, type ConfirmActionState } from '@/components/ui/confirm-action-dialog'
import { PageShell } from '@/components/layout/page-shell'
import { daysUntil, formatAmount } from '@/lib/utils'
import type { SupplierContact } from '@/api/suppliers'

export default function SupplierDetailPage() {
  const { id } = useParams({ strict: false })
  const supplierId = id!
  const apis = useApis()
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'

  const { data, loading, refresh } = useInjectedQuery({
    queryKey: queryKeys.suppliers.detail(supplierId),
    queryFn: (a) => a.suppliersApi.detail(supplierId),
    enabled: !!id,
  })

  const [contactOpen, setContactOpen] = useState(false)
  const [editingContact, setEditingContact] = useState<SupplierContact | null>(null)
  const [contactForm, setContactForm] = useState({
    name: '',
    position: '',
    phone: '',
    email: '',
    isPrimary: false,
  })
  const [contactSaving, setContactSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const openContactCreate = () => {
    setEditingContact(null)
    setContactForm({ name: '', position: '', phone: '', email: '', isPrimary: false })
    setContactOpen(true)
  }

  const openContactEdit = (c: SupplierContact) => {
    setEditingContact(c)
    setContactForm({
      name: c.name,
      position: c.position ?? '',
      phone: c.phone ?? '',
      email: c.email ?? '',
      isPrimary: c.isPrimary,
    })
    setContactOpen(true)
  }

  const handleContactSave = async () => {
    if (!contactForm.name) {
      toast.error('联系人姓名不能为空')
      return
    }
    setContactSaving(true)
    try {
      if (editingContact) {
        await apis.suppliersApi.updateContact(supplierId, editingContact.id, contactForm)
        toast.success('更新成功')
      } else {
        await apis.suppliersApi.createContact(supplierId, contactForm)
        toast.success('添加成功')
      }
      setContactOpen(false)
      refresh()
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setContactSaving(false)
    }
  }

  const handleContactDelete = (c: SupplierContact) => {
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除联系人「${c.name}」吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: async () => {
        await apis.suppliersApi.deleteContact(supplierId, c.id)
        toast.success('删除成功')
        refresh()
        setConfirmState(null)
      },
    })
  }

  if (loading) return <div className="p-6 text-muted-foreground">加载中...</div>
  if (!data) return <div className="p-6 text-muted-foreground">供应商不存在</div>

  const statusItem = SUPPLIER_STATUS[data.status]

  return (
    <PageShell>
      <Link
        to="/suppliers"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" /> 返回列表
      </Link>

      {/* 信息头卡 */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-semibold">{data.name}</h1>
                {statusItem && (
                  <span className="rounded-full border bg-muted/40 px-2 py-0.5 text-xs">
                    {statusItem.label}
                  </span>
                )}
              </div>
              <div className="mt-1 text-sm text-muted-foreground">
                编码：{data.code} · {data.category || '未分类'}
                {data.website && (
                  <a
                    href={data.website}
                    target="_blank"
                    rel="noreferrer"
                    className="ml-2 text-primary hover:underline"
                  >
                    官网
                  </a>
                )}
              </div>
              {data.description && (
                <p className="mt-2 text-sm text-muted-foreground">{data.description}</p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Tabs */}
      <Tabs defaultValue="contacts">
        <TabsList>
          <TabsTrigger value="contacts">联系人 ({data.contacts?.length ?? 0})</TabsTrigger>
          <TabsTrigger value="models">模型 ({data.models?.length ?? 0})</TabsTrigger>
          <TabsTrigger value="contracts">合同 ({data.contracts?.length ?? 0})</TabsTrigger>
          <TabsTrigger value="orders">订单 ({data.orders?.length ?? 0})</TabsTrigger>
          <TabsTrigger value="evaluations">评估历史 ({data.evaluations?.length ?? 0})</TabsTrigger>
        </TabsList>

        <TabsContent value="contacts" className="rounded-lg border bg-white p-4">
          {canEdit && (
            <div className="mb-3">
              <Button size="sm" onClick={openContactCreate}>
                <Plus className="h-3.5 w-3.5" /> 添加联系人
              </Button>
            </div>
          )}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>姓名</TableHead>
                <TableHead>职务</TableHead>
                <TableHead>电话</TableHead>
                <TableHead>邮箱</TableHead>
                {canEdit && <TableHead className="text-right">操作</TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.contacts.map((c) => (
                <TableRow key={c.id}>
                  <TableCell>
                    {c.name}
                    {c.isPrimary && (
                      <span className="ml-1 rounded bg-green-50 px-1.5 py-0.5 text-xs text-green-700">
                        主要
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{c.position || '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{c.phone || '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{c.email || '-'}</TableCell>
                  {canEdit && (
                    <TableCell className="text-right">
                      <Button variant="ghost" size="icon-sm" onClick={() => openContactEdit(c)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" onClick={() => handleContactDelete(c)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  )}
                </TableRow>
              ))}
              {data.contacts.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="py-8 text-center text-muted-foreground">
                    暂无联系人
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        <TabsContent value="models" className="rounded-lg border bg-white p-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>模型名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead className="text-right">上下文</TableHead>
                <TableHead className="text-right">输入价</TableHead>
                <TableHead className="text-right">输出价</TableHead>
                <TableHead className="text-right">折扣</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.models.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-medium">{m.modelName}</TableCell>
                  <TableCell className="text-muted-foreground">{m.modelType || '-'}</TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {m.contextLength ? `${(m.contextLength / 1000).toFixed(0)}K` : '-'}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {m.inputPrice ?? '-'}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {m.outputPrice ?? '-'}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {m.discount ? `${m.discount}%` : '-'}
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={m.status} map={MODEL_STATUS} />
                  </TableCell>
                </TableRow>
              ))}
              {data.models.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="py-8 text-center text-muted-foreground">
                    暂无模型
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        <TabsContent value="contracts" className="rounded-lg border bg-white p-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>合同编号</TableHead>
                <TableHead>标题</TableHead>
                <TableHead className="text-right">金额</TableHead>
                <TableHead>到期日</TableHead>
                <TableHead className="text-right">剩余</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.contracts.map((c) => {
                const days = daysUntil(c.endDate)
                return (
                  <TableRow key={c.id}>
                    <TableCell className="text-muted-foreground">{c.contractNo}</TableCell>
                    <TableCell>{c.title}</TableCell>
                    <TableCell className="text-right">{formatAmount(c.amount)}</TableCell>
                    <TableCell className="text-muted-foreground">{c.endDate ?? '-'}</TableCell>
                    <TableCell
                      className={`text-right text-xs font-medium ${days === null ? '' : days < 0 ? 'text-red-500' : days <= 30 ? 'text-yellow-600' : 'text-muted-foreground'}`}
                    >
                      {days === null ? '-' : days < 0 ? '已过期' : `${days} 天`}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={c.status} map={CONTRACT_STATUS} />
                    </TableCell>
                  </TableRow>
                )
              })}
              {data.contracts.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="py-8 text-center text-muted-foreground">
                    暂无合同
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        <TabsContent value="orders" className="rounded-lg border bg-white p-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>订单编号</TableHead>
                <TableHead className="text-right">金额</TableHead>
                <TableHead>下单日期</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>说明</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.orders.map((o) => (
                <TableRow key={o.id}>
                  <TableCell className="text-muted-foreground">{o.orderNo}</TableCell>
                  <TableCell className="text-right">{formatAmount(o.totalAmount)}</TableCell>
                  <TableCell className="text-muted-foreground">{o.orderDate ?? '-'}</TableCell>
                  <TableCell>
                    <StatusBadge status={o.status} map={ORDER_STATUS} />
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate text-muted-foreground">
                    {o.description ?? '-'}
                  </TableCell>
                </TableRow>
              ))}
              {data.orders.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="py-8 text-center text-muted-foreground">
                    暂无订单
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        <TabsContent value="evaluations" className="rounded-lg border bg-white p-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>周期</TableHead>
                <TableHead className="text-center">评级</TableHead>
                <TableHead className="text-right">综合分</TableHead>
                <TableHead className="text-right">质量</TableHead>
                <TableHead className="text-right">性能</TableHead>
                <TableHead className="text-right">价格</TableHead>
                <TableHead className="text-right">服务</TableHead>
                <TableHead className="text-right">合规</TableHead>
                <TableHead>评估人</TableHead>
                <TableHead>评语</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.evaluations.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="text-muted-foreground">{e.period}</TableCell>
                  <TableCell className="text-center">
                    <span
                      className={`inline-flex rounded px-2 py-0.5 text-xs font-bold text-white ${e.grade === 'A' ? 'bg-green-500' : e.grade === 'B' ? 'bg-blue-500' : e.grade === 'C' ? 'bg-yellow-500' : 'bg-red-500'}`}
                    >
                      {e.grade}
                    </span>
                  </TableCell>
                  <TableCell className="text-right font-medium">{e.totalScore}</TableCell>
                  <TableCell className="text-right text-muted-foreground">{e.quality}</TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {e.performance}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">{e.price}</TableCell>
                  <TableCell className="text-right text-muted-foreground">{e.service}</TableCell>
                  <TableCell className="text-right text-muted-foreground">{e.compliance}</TableCell>
                  <TableCell className="text-muted-foreground">{e.evaluatorName ?? '-'}</TableCell>
                  <TableCell className="max-w-[160px] truncate text-muted-foreground">
                    {e.comment ?? '-'}
                  </TableCell>
                </TableRow>
              ))}
              {data.evaluations.length === 0 && (
                <TableRow>
                  <TableCell colSpan={10} className="py-8 text-center text-muted-foreground">
                    暂无评估记录
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>
      </Tabs>

      {/* 联系人弹窗 */}
      <Dialog open={contactOpen} onOpenChange={setContactOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editingContact ? '编辑联系人' : '添加联系人'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>
                姓名 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={contactForm.name}
                onChange={(e) => setContactForm({ ...contactForm, name: e.target.value })}
              />
            </div>
            <div>
              <Label>职务</Label>
              <Input
                className="mt-1"
                value={contactForm.position}
                onChange={(e) => setContactForm({ ...contactForm, position: e.target.value })}
              />
            </div>
            <div>
              <Label>电话</Label>
              <Input
                className="mt-1"
                value={contactForm.phone}
                onChange={(e) => setContactForm({ ...contactForm, phone: e.target.value })}
              />
            </div>
            <div>
              <Label>邮箱</Label>
              <Input
                className="mt-1"
                value={contactForm.email}
                onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="isPrimary"
                checked={contactForm.isPrimary}
                onChange={(e) => setContactForm({ ...contactForm, isPrimary: e.target.checked })}
              />
              <Label htmlFor="isPrimary">主要联系人</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setContactOpen(false)}>
              取消
            </Button>
            <Button onClick={handleContactSave} disabled={contactSaving}>
              {contactSaving ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmActionDialog
        state={confirmState}
        onOpenChange={(open) => !open && setConfirmState(null)}
        onClose={() => setConfirmState(null)}
      />
    </PageShell>
  )
}
