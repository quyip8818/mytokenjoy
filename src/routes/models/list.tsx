import { toast } from 'sonner'
import { Box } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { modelApi } from '@/api/models'
import type { ModelInfo } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { listEmpty } from '@/lib/list-empty'
import { useRowHighlight } from '@/lib/use-row-highlight'
import { PROVIDER_CHIP_STYLES, PROVIDER_LABELS } from '@/lib/labels'
import { StatusBadge } from '@/components/ui/status-badge'

export default function ModelListPage() {
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useDemoCta('MODEL')
  const { data: models = [], loading, refresh } = useAsyncResource(() => modelApi.list(), [])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

  const handleToggle = async (model: ModelInfo) => {
    await modelApi.toggle(model.id, !model.enabled)
    toast.success(model.enabled ? '模型已禁用' : '模型已启用')
    flashRow(model.id)
    void refresh()
  }

  const openCreate = () => openWithRefresh('model-create')

  return (
    <PageShell
      actions={
        <Button
          id={modelCta.id}
          size="sm"
          variant="brand"
          className={modelCta.className}
          onClick={openCreate}
        >
          添加模型
        </Button>
      }
    >
      <DataSection
        loading={loading}
        skeletonColumns={7}
        empty={listEmpty(loading, models, {
          icon: Box,
          title: '暂无模型',
          description: '添加自定义模型以扩展可用模型列表',
          actionLabel: '添加模型',
          onAction: openCreate,
        })}
      >
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>模型名称</TableHead>
              <TableHead>供应商</TableHead>
              <TableHead className="text-right">输入价格</TableHead>
              <TableHead className="text-right">输出价格</TableHead>
              <TableHead className="text-right">上下文窗口</TableHead>
              <TableHead>能力</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {models.map((model) => (
              <TableRow
                key={model.id}
                className={`${rowClass(model.id)} ${!model.enabled ? 'opacity-50' : ''}`}
              >
                <TableCell className="font-medium">{model.displayName}</TableCell>
                <TableCell>
                  <Badge
                    variant="outline"
                    className={`border-0 ${PROVIDER_CHIP_STYLES[model.provider] ?? PROVIDER_CHIP_STYLES.custom}`}
                  >
                    {PROVIDER_LABELS[model.provider]}
                  </Badge>
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  ¥{model.inputPrice}/M
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  ¥{model.outputPrice}/M
                </TableCell>
                <TableCell className="text-right text-sm">
                  {(model.maxContext / 1000).toFixed(0)}K
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {model.capabilities.map((c) => (
                      <StatusBadge key={c} variant="neutral" className="text-[11px]">
                        {c}
                      </StatusBadge>
                    ))}
                  </div>
                </TableCell>
                <TableCell>
                  <Switch checked={model.enabled} onCheckedChange={() => handleToggle(model)} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
