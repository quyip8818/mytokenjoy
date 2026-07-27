import { useState, useEffect, useMemo } from 'react'
import { toast } from 'sonner'
import { useApis } from '@/api/use-apis'
import { DIMENSIONS } from '@/config/enums'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { EvaluationWeight } from '@/api/evaluations'

const DIMENSION_DESC: Record<string, string> = {
  quality: '模型输出的准确性、任务完成质量',
  performance: '响应延迟、服务可用性与稳定性',
  price: '单价水平与折扣力度、性价比',
  service: '技术支持响应与问题解决能力',
  compliance: '数据安全、合规资质与内容安全',
}

export default function WeightsPage() {
  const apis = useApis()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [weights, setWeights] = useState<EvaluationWeight[]>([])

  useEffect(() => {
    let cancelled = false
    async function fetchWeights() {
      setLoading(true)
      try {
        const data = await apis.evaluationsApi.getWeights()
        if (!cancelled) setWeights(data)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    fetchWeights()
    return () => {
      cancelled = true
    }
  }, [apis])

  const totalWeight = useMemo(() => weights.reduce((sum, w) => sum + w.weight, 0), [weights])
  const isValid = Math.abs(totalWeight - 100) < 0.01

  const updateWeight = (dimension: string, value: number) => {
    setWeights((prev) => prev.map((w) => (w.dimension === dimension ? { ...w, weight: value } : w)))
  }

  const handleSave = async () => {
    if (!isValid) {
      toast.error('五项权重合计必须等于 100')
      return
    }
    setSaving(true)
    try {
      await apis.evaluationsApi.updateWeights(weights)
      toast.success('权重配置已更新')
      const data = await apis.evaluationsApi.getWeights()
      setWeights(data)
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <PageShell>
      <PageHeader
        title="绩效评估权重配置"
        description="配置五个评估维度在综合分中的权重占比，合计必须等于 100%"
        actions={
          <button
            onClick={handleSave}
            disabled={saving || !isValid}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
          >
            {saving ? '保存中...' : '保存配置'}
          </button>
        }
      />

      <div className="mx-auto max-w-2xl rounded-lg border bg-white p-6">
        {loading ? (
          <div className="py-8 text-center text-muted-foreground">加载中...</div>
        ) : (
          <div className="space-y-1">
            {weights.map((w) => (
              <div
                key={w.dimension}
                className="flex items-center justify-between border-b border-dashed py-4 last:border-0"
              >
                <div className="min-w-[140px]">
                  <div className="text-sm font-medium">
                    {DIMENSIONS[w.dimension] ?? w.dimension}
                  </div>
                  <div className="text-xs text-muted-foreground">{DIMENSION_DESC[w.dimension]}</div>
                </div>
                <div className="flex items-center gap-3">
                  <input
                    type="range"
                    min={0}
                    max={100}
                    step={5}
                    value={w.weight}
                    onChange={(e) => updateWeight(w.dimension, Number(e.target.value))}
                    className="w-48"
                  />
                  <input
                    type="number"
                    min={0}
                    max={100}
                    className="w-16 rounded border px-2 py-1 text-center text-sm"
                    value={w.weight}
                    onChange={(e) => updateWeight(w.dimension, Number(e.target.value))}
                  />
                  <span className="w-5 text-sm text-muted-foreground">%</span>
                </div>
              </div>
            ))}
          </div>
        )}

        <div
          className={`mt-4 rounded-md p-3 text-sm font-medium ${isValid ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}
        >
          当前合计：{totalWeight}%{isValid ? ' ✓ 可保存' : ' （需调整至 100% 才能保存）'}
        </div>
      </div>
    </PageShell>
  )
}
