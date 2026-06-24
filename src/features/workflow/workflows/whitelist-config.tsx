import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { routingApi } from '@/api/models'
import { departmentApi } from '@/api/org'
import type { RoutingRule } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useWorkflow } from '../use-workflow'

import { findParentDeptId } from '@/lib/org'

export function WhitelistConfigWorkflow({
  entry,
  onClose,
  onPush,
  onSetDirty,
}: WorkflowComponentProps<'whitelist-config'>) {
  const { closeAll } = useWorkflow()
  const rule = entry.payload.rule as RoutingRule
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [inherited, setInherited] = useState(rule.inherited)
  const [models, setModels] = useState<string[]>(rule.allowedModels)
  const [parentWhitelist, setParentWhitelist] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    void (async () => {
      const [rules, depts] = await Promise.all([routingApi.getRules(), departmentApi.getTree()])
      const parentId = findParentDeptId(depts, rule.nodeId)
      const parentRule = parentId ? rules.find((r) => r.nodeId === parentId) : undefined
      setParentWhitelist(parentRule?.allowedModels ?? rule.allowedModels)
    })()
  }, [rule.nodeId, rule.allowedModels])

  const handlePickModels = () => {
    onPush('model-picker', {
      selectedModels: models,
      parentWhitelist,
      onConfirm: (picked: string[]) => {
        setModels(picked)
        onSetDirty(true)
      },
    })
  }

  const handleSave = async () => {
    if (!inherited && models.length === 0) {
      toast.error('请至少选择一个模型')
      return
    }
    setSubmitting(true)
    try {
      await routingApi.updateRule(rule.id, {
        inherited,
        allowedModels: inherited ? rule.allowedModels : models,
      })
      toast.success('白名单已保存')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="配置部门白名单"
      onClose={onClose}
      contextBar={rule.nodeName}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSave}
          primaryDisabled={submitting}
        />
      }
    >
      <div className="grid grid-cols-5 gap-8">
        <div className="col-span-3 space-y-5">
          <div className="flex items-center justify-between rounded-lg border border-border/50 px-4 py-3">
            <div>
              <Label>继承父级白名单</Label>
              <p className="text-xs text-muted-foreground mt-1">
                开启后使用父级配置，关闭可自定义勾选
              </p>
            </div>
            <Switch
              checked={inherited}
              onCheckedChange={(checked) => {
                setInherited(checked)
                onSetDirty(true)
              }}
            />
          </div>
          {!inherited && (
            <div className="space-y-3">
              <Label>允许模型</Label>
              <Button variant="outline" onClick={handlePickModels}>
                选择模型 ({models.length})
              </Button>
              <div className="flex flex-wrap gap-1">
                {models.map((m) => (
                  <Badge key={m} variant="outline" className="text-xs">
                    {m}
                  </Badge>
                ))}
              </div>
            </div>
          )}
          {inherited && (
            <p className="text-sm text-muted-foreground">
              当前继承父级，共 {parentWhitelist.length} 个可用模型
            </p>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-2">
          <h4 className="font-semibold">父级白名单参考</h4>
          <div className="flex flex-wrap gap-1">
            {parentWhitelist.map((m) => (
              <Badge key={m} variant="outline" className="text-xs">
                {m}
              </Badge>
            ))}
          </div>
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}
