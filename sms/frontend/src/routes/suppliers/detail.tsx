import { useState } from 'react'
import { useParams, Link } from 'react-router'
import { ArrowLeft, Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useInjectedQuery, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'
import { SUPPLIER_STATUS, CONTRACT_STATUS, ORDER_STATUS, MODEL_STATUS } from '@/config/enums'
import { StatusBadge } from '@/components/ui'
import { PageShell } from '@/components/layout/page-shell'
import type { SupplierContact } from '@/api/suppliers'

function daysUntil(endDate?: string): number | null {
  if (!endDate) return null
  const end = new Date(endDate).getTime()
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return Math.ceil((end - now.getTime()) / (24 * 3600 * 1000))
}

function formatAmount(amount?: number): string {
  if (amount === undefined || amount === null) return '-'
  return Number(amount).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

export default function SupplierDetailPage() {
  const { id } = useParams()
  const supplierId = Number(id)
  const apis = useApis()
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'

  const { data, loading, refresh } = useInjectedQuery({
    queryKey: queryKeys.suppliers.detail(supplierId),
    queryFn: (a) => a.suppliersApi.detail(supplierId),
    enabled: !!id,
  })

  const [tab, setTab] = useState('contacts')
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

  const handleContactDelete = async (c: SupplierContact) => {
    if (!confirm(`确定删除联系人「${c.name}」吗？`)) return
    await apis.suppliersApi.deleteContact(supplierId, c.id)
    toast.success('删除成功')
    refresh()
  }

  if (loading) return <div className="p-6 text-muted-foreground">加载中...</div>
  if (!data) return <div className="p-6 text-muted-foreground">供应商不存在</div>

  const statusItem = SUPPLIER_STATUS[data.status]
  const tabs = [
    { key: 'contacts', label: `联系人 (${data.contacts?.length ?? 0})` },
    { key: 'models', label: `模型 (${data.models?.length ?? 0})` },
    { key: 'contracts', label: `合同 (${data.contracts?.length ?? 0})` },
    { key: 'orders', label: `订单 (${data.orders?.length ?? 0})` },
    { key: 'evaluations', label: `评估历史 (${data.evaluations?.length ?? 0})` },
  ]

  return (
    <PageShell>
      <Link
        to="/suppliers"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" /> 返回列表
      </Link>

      {/* 信息头卡 */}
      <div className="rounded-lg border bg-white p-6">
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
      </div>

      {/* Tab */}
      <div className="rounded-lg border bg-white">
        <div className="flex border-b">
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`px-4 py-3 text-sm font-medium border-b-2 -mb-px ${tab === t.key ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <div className="p-4">
          {tab === 'contacts' && (
            <div>
              {canEdit && (
                <div className="mb-3">
                  <button
                    onClick={openContactCreate}
                    className="inline-flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground"
                  >
                    <Plus className="h-3.5 w-3.5" /> 添加联系人
                  </button>
                </div>
              )}
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/30">
                  <tr>
                    <th className="px-3 py-2 text-left font-medium">姓名</th>
                    <th className="px-3 py-2 text-left font-medium">职务</th>
                    <th className="px-3 py-2 text-left font-medium">电话</th>
                    <th className="px-3 py-2 text-left font-medium">邮箱</th>
                    {canEdit && <th className="px-3 py-2 text-right font-medium">操作</th>}
                  </tr>
                </thead>
                <tbody>
                  {data.contacts.map((c) => (
                    <tr key={c.id} className="border-b last:border-0">
                      <td className="px-3 py-2">
                        {c.name}
                        {c.isPrimary && (
                          <span className="ml-1 rounded bg-green-50 px-1.5 py-0.5 text-xs text-green-700">
                            主要
                          </span>
                        )}
                      </td>
                      <td className="px-3 py-2 text-muted-foreground">{c.position || '-'}</td>
                      <td className="px-3 py-2 text-muted-foreground">{c.phone || '-'}</td>
                      <td className="px-3 py-2 text-muted-foreground">{c.email || '-'}</td>
                      {canEdit && (
                        <td className="px-3 py-2 text-right">
                          <button
                            onClick={() => openContactEdit(c)}
                            className="p-1 text-muted-foreground hover:text-primary"
                          >
                            <Pencil className="h-3.5 w-3.5" />
                          </button>
                          <button
                            onClick={() => handleContactDelete(c)}
                            className="p-1 text-muted-foreground hover:text-red-500"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        </td>
                      )}
                    </tr>
                  ))}
                  {data.contacts.length === 0 && (
                    <tr>
                      <td colSpan={5} className="py-8 text-center text-muted-foreground">
                        暂无联系人
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}

          {tab === 'models' && (
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/30">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">模型名称</th>
                  <th className="px-3 py-2 text-left font-medium">类型</th>
                  <th className="px-3 py-2 text-right font-medium">上下文</th>
                  <th className="px-3 py-2 text-right font-medium">输入价</th>
                  <th className="px-3 py-2 text-right font-medium">输出价</th>
                  <th className="px-3 py-2 text-right font-medium">折扣</th>
                  <th className="px-3 py-2 text-left font-medium">状态</th>
                </tr>
              </thead>
              <tbody>
                {data.models.map((m) => (
                  <tr key={m.id} className="border-b last:border-0">
                    <td className="px-3 py-2 font-medium">{m.modelName}</td>
                    <td className="px-3 py-2 text-muted-foreground">{m.modelType || '-'}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">
                      {m.contextLength ? `${(m.contextLength / 1000).toFixed(0)}K` : '-'}
                    </td>
                    <td className="px-3 py-2 text-right text-muted-foreground">
                      {m.inputPrice ?? '-'}
                    </td>
                    <td className="px-3 py-2 text-right text-muted-foreground">
                      {m.outputPrice ?? '-'}
                    </td>
                    <td className="px-3 py-2 text-right text-muted-foreground">
                      {m.discount ? `${m.discount}%` : '-'}
                    </td>
                    <td className="px-3 py-2">
                      <StatusBadge status={m.status} map={MODEL_STATUS} />
                    </td>
                  </tr>
                ))}
                {data.models.length === 0 && (
                  <tr>
                    <td colSpan={7} className="py-8 text-center text-muted-foreground">
                      暂无模型
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}

          {tab === 'contracts' && (
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/30">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">合同编号</th>
                  <th className="px-3 py-2 text-left font-medium">标题</th>
                  <th className="px-3 py-2 text-right font-medium">金额</th>
                  <th className="px-3 py-2 text-left font-medium">到期日</th>
                  <th className="px-3 py-2 text-right font-medium">剩余</th>
                  <th className="px-3 py-2 text-left font-medium">状态</th>
                </tr>
              </thead>
              <tbody>
                {data.contracts.map((c) => {
                  const days = daysUntil(c.endDate)
                  return (
                    <tr key={c.id} className="border-b last:border-0">
                      <td className="px-3 py-2 text-muted-foreground">{c.contractNo}</td>
                      <td className="px-3 py-2">{c.title}</td>
                      <td className="px-3 py-2 text-right">{formatAmount(c.amount)}</td>
                      <td className="px-3 py-2 text-muted-foreground">{c.endDate ?? '-'}</td>
                      <td
                        className={`px-3 py-2 text-right text-xs font-medium ${days === null ? '' : days < 0 ? 'text-red-500' : days <= 30 ? 'text-yellow-600' : 'text-muted-foreground'}`}
                      >
                        {days === null ? '-' : days < 0 ? '已过期' : `${days} 天`}
                      </td>
                      <td className="px-3 py-2">
                        <StatusBadge status={c.status} map={CONTRACT_STATUS} />
                      </td>
                    </tr>
                  )
                })}
                {data.contracts.length === 0 && (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-muted-foreground">
                      暂无合同
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}

          {tab === 'orders' && (
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/30">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">订单编号</th>
                  <th className="px-3 py-2 text-right font-medium">金额</th>
                  <th className="px-3 py-2 text-left font-medium">下单日期</th>
                  <th className="px-3 py-2 text-left font-medium">状态</th>
                  <th className="px-3 py-2 text-left font-medium">说明</th>
                </tr>
              </thead>
              <tbody>
                {data.orders.map((o) => (
                  <tr key={o.id} className="border-b last:border-0">
                    <td className="px-3 py-2 text-muted-foreground">{o.orderNo}</td>
                    <td className="px-3 py-2 text-right">{formatAmount(o.totalAmount)}</td>
                    <td className="px-3 py-2 text-muted-foreground">{o.orderDate ?? '-'}</td>
                    <td className="px-3 py-2">
                      <StatusBadge status={o.status} map={ORDER_STATUS} />
                    </td>
                    <td className="px-3 py-2 text-muted-foreground truncate max-w-[200px]">
                      {o.description ?? '-'}
                    </td>
                  </tr>
                ))}
                {data.orders.length === 0 && (
                  <tr>
                    <td colSpan={5} className="py-8 text-center text-muted-foreground">
                      暂无订单
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}

          {tab === 'evaluations' && (
            <table className="w-full text-sm">
              <thead className="border-b bg-muted/30">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">周期</th>
                  <th className="px-3 py-2 text-center font-medium">评级</th>
                  <th className="px-3 py-2 text-right font-medium">综合分</th>
                  <th className="px-3 py-2 text-right font-medium">质量</th>
                  <th className="px-3 py-2 text-right font-medium">性能</th>
                  <th className="px-3 py-2 text-right font-medium">价格</th>
                  <th className="px-3 py-2 text-right font-medium">服务</th>
                  <th className="px-3 py-2 text-right font-medium">合规</th>
                  <th className="px-3 py-2 text-left font-medium">评估人</th>
                  <th className="px-3 py-2 text-left font-medium">评语</th>
                </tr>
              </thead>
              <tbody>
                {data.evaluations.map((e) => (
                  <tr key={e.id} className="border-b last:border-0">
                    <td className="px-3 py-2 text-muted-foreground">{e.period}</td>
                    <td className="px-3 py-2 text-center">
                      <span
                        className={`inline-flex rounded px-2 py-0.5 text-xs font-bold text-white ${e.grade === 'A' ? 'bg-green-500' : e.grade === 'B' ? 'bg-blue-500' : e.grade === 'C' ? 'bg-yellow-500' : 'bg-red-500'}`}
                      >
                        {e.grade}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-right font-medium">{e.totalScore}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">{e.quality}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">{e.performance}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">{e.price}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">{e.service}</td>
                    <td className="px-3 py-2 text-right text-muted-foreground">{e.compliance}</td>
                    <td className="px-3 py-2 text-muted-foreground">{e.evaluatorName ?? '-'}</td>
                    <td className="px-3 py-2 text-muted-foreground truncate max-w-[160px]">
                      {e.comment ?? '-'}
                    </td>
                  </tr>
                ))}
                {data.evaluations.length === 0 && (
                  <tr>
                    <td colSpan={10} className="py-8 text-center text-muted-foreground">
                      暂无评估记录
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* 联系人弹窗 */}
      {contactOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setContactOpen(false)}
        >
          <div
            className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">
              {editingContact ? '编辑联系人' : '添加联系人'}
            </h2>
            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-sm font-medium">
                  姓名 <span className="text-red-500">*</span>
                </label>
                <input
                  className="input"
                  value={contactForm.name}
                  onChange={(e) => setContactForm({ ...contactForm, name: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">职务</label>
                <input
                  className="input"
                  value={contactForm.position}
                  onChange={(e) => setContactForm({ ...contactForm, position: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">电话</label>
                <input
                  className="input"
                  value={contactForm.phone}
                  onChange={(e) => setContactForm({ ...contactForm, phone: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">邮箱</label>
                <input
                  className="input"
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
                <label htmlFor="isPrimary" className="text-sm">
                  主要联系人
                </label>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <button
                onClick={() => setContactOpen(false)}
                className="rounded-md border px-4 py-2 text-sm"
              >
                取消
              </button>
              <button
                onClick={handleContactSave}
                disabled={contactSaving}
                className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
              >
                {contactSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </PageShell>
  )
}
