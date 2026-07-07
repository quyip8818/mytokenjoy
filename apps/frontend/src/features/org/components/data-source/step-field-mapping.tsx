import { useEffect, useState } from 'react'
import type { FieldMapping, MappingTestResult, Platform } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ArrowRight, CheckCircle2, Loader2, Search } from 'lucide-react'

const targetFields = [
  { value: 'name', label: '姓名' },
  { value: 'phone', label: '手机号' },
  { value: 'email', label: '邮箱' },
  { value: 'departmentName', label: '部门名称' },
  { value: 'departmentId', label: '部门 ID' },
  { value: 'status', label: '状态' },
  { value: 'roles', label: '角色' },
  { value: '_ignore', label: '（忽略）' },
]

interface StepFieldMappingProps {
  platform: Platform
  dataSourceApi: AppApis['dataSourceApi']
  onComplete: () => void
  onBack: () => void
}

export function StepFieldMapping({
  platform,
  dataSourceApi,
  onComplete,
  onBack,
}: StepFieldMappingProps) {
  const [mappings, setMappings] = useState<FieldMapping[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testKeyword, setTestKeyword] = useState('')
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<MappingTestResult | null>(null)
  const [testPassed, setTestPassed] = useState(false)

  useEffect(() => {
    void dataSourceApi.getFieldMappings(platform).then((data) => {
      setMappings(data)
      setLoading(false)
    })
  }, [dataSourceApi, platform])

  const updateMapping = (index: number, targetField: string) => {
    setMappings((prev) => prev.map((m, i) => (i === index ? { ...m, targetField } : m)))
    setTestPassed(false)
    setTestResult(null)
  }

  const handleTest = async () => {
    if (!testKeyword.trim()) return
    setTesting(true)
    setTestResult(null)
    try {
      const result = await dataSourceApi.testFieldMapping(platform, testKeyword)
      setTestResult(result)
      if (result.success) setTestPassed(true)
    } finally {
      setTesting(false)
    }
  }

  const handleSaveAndNext = async () => {
    setSaving(true)
    try {
      await dataSourceApi.saveFieldMappings({ platform, mappings })
      onComplete()
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-semibold">字段映射配置</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          将源系统的字段映射到本系统对应字段，配置完成后测试验证映射结果
        </p>
      </div>

      {/* Mapping table */}
      <div className="rounded-lg border">
        <div className="grid grid-cols-[1fr_32px_1fr] gap-2 border-b bg-muted/50 px-4 py-2.5 text-xs font-medium text-muted-foreground">
          <span>源字段</span>
          <span />
          <span>目标字段</span>
        </div>
        <div className="divide-y">
          {mappings.map((mapping, index) => (
            <div
              key={mapping.sourceField}
              className="grid grid-cols-[1fr_32px_1fr] items-center gap-2 px-4 py-3"
            >
              <div className="flex items-center gap-2">
                <span className="text-sm">{mapping.sourceLabel}</span>
                {mapping.required && <span className="text-xs text-destructive">*</span>}
              </div>
              <ArrowRight className="size-4 text-muted-foreground/50 justify-self-center" />
              <Select
                value={mapping.targetField}
                onValueChange={(val) => updateMapping(index, val)}
              >
                <SelectTrigger size="sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {targetFields.map((f) => (
                    <SelectItem key={f.value} value={f.value}>
                      {f.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          ))}
        </div>
      </div>

      {/* Test area */}
      <div className="rounded-lg border border-border bg-muted/30 p-4 space-y-3">
        <Label className="text-sm font-medium">测试映射</Label>
        <div className="flex gap-2 max-w-sm">
          <Input
            value={testKeyword}
            onChange={(e) => setTestKeyword(e.target.value)}
            placeholder="输入姓名搜索测试"
            onKeyDown={(e) => e.key === 'Enter' && handleTest()}
          />
          <Button
            variant="outline"
            size="default"
            onClick={handleTest}
            disabled={testing || !testKeyword.trim()}
          >
            {testing ? <Loader2 className="size-4 animate-spin" /> : <Search className="size-4" />}
            测试
          </Button>
        </div>

        {testResult && (
          <div
            className={`rounded-md border px-4 py-3 text-sm ${
              testResult.success ? 'border-emerald-200 bg-emerald-50' : 'border-red-200 bg-red-50'
            }`}
          >
            {testResult.success ? (
              <div className="space-y-1.5">
                <p className="flex items-center gap-1.5 font-medium text-emerald-700">
                  <CheckCircle2 className="size-4" />
                  映射测试通过
                </p>
                <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-emerald-700 mt-2">
                  {Object.entries(testResult.preview).map(([key, value]) => (
                    <div key={key} className="flex gap-1.5">
                      <span className="text-muted-foreground">{key}:</span>
                      <span className="font-medium text-foreground">{value}</span>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="space-y-1 text-destructive">
                <p className="font-medium">映射测试失败</p>
                {testResult.errors.map((err, i) => (
                  <p key={i} className="text-xs">
                    {err}
                  </p>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-3 pt-2">
        <Button variant="outline" onClick={onBack}>
          上一步
        </Button>
        <Button onClick={handleSaveAndNext} disabled={!testPassed || saving}>
          {saving && <Loader2 className="size-4 animate-spin" />}
          下一步
        </Button>
      </div>
    </div>
  )
}
