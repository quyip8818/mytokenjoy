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
import { listEmpty } from '@/lib/list-empty'
import { PROVIDER_CHIP_STYLES, PROVIDER_LABELS } from '@/lib/labels'
import { StatusBadge } from '@/components/ui/status-badge'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { useModelListPage } from '@/routes/models/hooks/use-model-list-page'

export default function ModelListPage() {
  const {
    models,
    loading,
    error,
    refresh,
    canManage,
    modelCta,
    rowClass,
    handleToggle,
    openCreate,
  } = useModelListPage()

  return (
    <PageShell
      actions={
        <PermissionGate write permission={PERMISSION.MODEL_MANAGE}>
          <Button
            id={modelCta.id}
            size="sm"
            variant="brand"
            className={modelCta.className}
            onClick={openCreate}
          >
            添加模型
          </Button>
        </PermissionGate>
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={7}
        empty={listEmpty(loading, models, {
          icon: Box,
          title: '暂无模型',
          description: '添加自定义模型以扩展可用模型列表',
          actionLabel: canManage ? '添加模型' : undefined,
          onAction: canManage ? openCreate : undefined,
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
                  {canManage ? (
                    <Switch checked={model.enabled} onCheckedChange={() => handleToggle(model)} />
                  ) : (
                    <StatusBadge variant={model.enabled ? 'success' : 'neutral'}>
                      {model.enabled ? '启用' : '禁用'}
                    </StatusBadge>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
