import { Pencil, Power, Trash2, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { PlatformModel } from '@/api/platform'
import type { usePlatformModelsPage } from '../hooks/use-platform-models-page'

type Props = ReturnType<typeof usePlatformModelsPage>

function PricingDialog({
  model,
  form,
  setForm,
  onSave,
  onClose,
}: {
  model: PlatformModel
  form: { inputPrice: string; outputPrice: string }
  setForm: (f: { inputPrice: string; outputPrice: string }) => void
  onSave: () => void
  onClose: () => void
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-base font-semibold">编辑定价</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          {model.name} ({model.type})
        </p>
        <div className="mt-4 space-y-3">
          <label className="block text-sm">
            <span className="text-muted-foreground">输入价格 (元/百万tokens)</span>
            <input
              type="number"
              className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
              value={form.inputPrice}
              onChange={(e) => setForm({ ...form, inputPrice: e.target.value })}
            />
          </label>
          <label className="block text-sm">
            <span className="text-muted-foreground">输出价格 (元/百万tokens)</span>
            <input
              type="number"
              className="mt-1 w-full rounded-md border px-3 py-2 text-sm"
              value={form.outputPrice}
              onChange={(e) => setForm({ ...form, outputPrice: e.target.value })}
            />
          </label>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onClose}>
            取消
          </Button>
          <Button size="sm" onClick={onSave}>
            保存
          </Button>
        </div>
      </div>
    </div>
  )
}

export function PlatformModelsPageShell({
  models,
  loading,
  error,
  refresh,
  publishing,
  handlePublish,
  handleToggle,
  handleDelete,
  pricingModel,
  pricingForm,
  setPricingForm,
  openPricing,
  closePricing,
  handleSavePricing,
}: Props) {
  return (
    <PageShell>
      <PageHeader
        title="模型目录"
        description={`共 ${models.length} 个全局模型`}
        actions={
          <Button size="sm" variant="brand" disabled={publishing} onClick={handlePublish}>
            <Upload className="mr-1.5 h-4 w-4" />
            {publishing ? '发布中...' : '发布'}
          </Button>
        }
      />

      <Card className="border-border shadow-xs">
        <CardContent className="px-5 pt-5 pb-4">
          <DataSection loading={loading} error={error} onRetry={refresh} skeletonColumns={6}>
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/30 text-xs text-muted-foreground">
                  <th className="px-4 py-3 text-left font-medium">模型</th>
                  <th className="px-4 py-3 text-left font-medium">Provider</th>
                  <th className="px-4 py-3 text-right font-medium">输入价格</th>
                  <th className="px-4 py-3 text-right font-medium">输出价格</th>
                  <th className="px-4 py-3 text-center font-medium">状态</th>
                  <th className="px-4 py-3 text-center font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {models.map((m) => (
                  <tr
                    key={m.modelId}
                    className={`border-b last:border-0 hover:bg-muted/20 ${!m.active ? 'opacity-50' : ''}`}
                  >
                    <td className="px-4 py-3">
                      <div className="font-medium">{m.name}</div>
                      <div className="font-mono text-[11px] text-muted-foreground">{m.type}</div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{m.provider}</td>
                    <td className="px-4 py-3 text-right font-mono text-xs">
                      {m.inputPrice > 0 ? `¥${m.inputPrice}` : '—'}
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-xs">
                      {m.outputPrice > 0 ? `¥${m.outputPrice}` : '—'}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <Badge variant={m.active ? 'default' : 'outline'}>
                        {m.active ? '启用' : '禁用'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <div className="inline-flex items-center gap-1">
                        <button
                          onClick={() => openPricing(m)}
                          className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                          title="编辑定价"
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => handleToggle(m)}
                          className={`rounded p-1.5 hover:bg-muted ${m.active ? 'text-amber-500' : 'text-green-500'}`}
                          title={m.active ? '禁用' : '启用'}
                        >
                          <Power className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(m)}
                          className="rounded p-1.5 text-muted-foreground hover:bg-red-50 hover:text-red-600"
                          title="删除"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </DataSection>
        </CardContent>
      </Card>

      {pricingModel && (
        <PricingDialog
          model={pricingModel}
          form={pricingForm}
          setForm={setPricingForm}
          onSave={handleSavePricing}
          onClose={closePricing}
        />
      )}
    </PageShell>
  )
}
