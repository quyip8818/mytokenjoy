import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Box } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
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
import { modelApi } from '@/api/models'
import type { ModelInfo } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoCta } from '@/features/demo'
import { EmptyState } from '@/components/ui/empty-state'
import { useRowHighlight } from '@/lib/use-row-highlight'
import { cn } from '@/lib/utils'

import { PROVIDER_CHIP_STYLES, PROVIDER_LABELS } from '@/lib/labels'

export default function ModelListPage() {
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const modelCta = useDemoCta('MODEL')
  const [models, setModels] = useState<ModelInfo[]>([])

  const load = useCallback(async () => {
    const list = await modelApi.list()
    setModels(list)
  }, [])

  useEffect(() => {
    void modelApi.list().then(setModels)
  }, [])

  const handleToggle = async (model: ModelInfo) => {
    await modelApi.toggle(model.id, !model.enabled)
    toast.success(model.enabled ? '模型已禁用' : '模型已启用')
    flashRow(model.id)
    void load()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <Button
          id={modelCta.id}
          size="sm"
          className={cn(
            'bg-gradient-to-r from-indigo-600 to-violet-600 text-white',
            modelCta.className,
          )}
          onClick={() =>
            open('model-create', {
              onSuccess: (id?: string) => {
                void load()
                if (id) flashRow(id)
              },
            })
          }
        >
          添加模型
        </Button>
      </div>
      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          {models.length === 0 ? (
            <EmptyState
              icon={Box}
              title="暂无模型"
              description="添加自定义模型以扩展可用模型列表"
              actionLabel="添加模型"
              onAction={() =>
                open('model-create', {
                  onSuccess: (id?: string) => {
                    void load()
                    if (id) flashRow(id)
                  },
                })
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    模型名称
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    供应商
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                    输入价格
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                    输出价格
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                    上下文窗口
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    能力
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    状态
                  </TableHead>
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
                          <Badge
                            key={c}
                            variant="secondary"
                            className="bg-slate-100 text-slate-600 text-[11px] border-0"
                          >
                            {c}
                          </Badge>
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
          )}
        </CardContent>
      </Card>
    </div>
  )
}
