import { useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useInjectedApis } from '@/api/use-apis'
import type { ChannelSummary } from '@/api/types'
import type { WorkflowComponentProps } from '../types'

export function PlatformModelChannelsWorkflow({
  entry,
}: WorkflowComponentProps<'platform-model-channels'>) {
  const { model } = entry.payload
  const apis = useInjectedApis()
  const [channels, setChannels] = useState<ChannelSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    apis.platformApi
      .listModelChannels(model.modelId)
      .then((data) => {
        if (!cancelled) setChannels(data)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : '加载失败')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [apis, model.modelId])

  return (
    <div className="flex flex-col gap-4 p-4">
      <div>
        <h3 className="text-sm font-medium text-foreground">{model.name}</h3>
        <p className="text-xs text-muted-foreground font-mono">{model.type}</p>
      </div>

      {loading && <p className="text-sm text-muted-foreground">加载中...</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}

      {!loading && !error && channels.length === 0 && (
        <p className="text-sm text-muted-foreground">该模型没有关联的渠道</p>
      )}

      {!loading && channels.length > 0 && (
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent border-border/60">
              <TableHead className="text-xs font-medium text-muted-foreground">渠道名</TableHead>
              <TableHead className="text-xs font-medium text-muted-foreground">Group</TableHead>
              <TableHead className="text-xs font-medium text-muted-foreground text-right">
                优先级
              </TableHead>
              <TableHead className="text-xs font-medium text-muted-foreground text-right">
                权重
              </TableHead>
              <TableHead className="text-xs font-medium text-muted-foreground">状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {channels.map((ch) => (
              <TableRow key={ch.name} className="hover:bg-muted/30">
                <TableCell className="text-sm font-medium">{ch.name}</TableCell>
                <TableCell className="text-sm text-muted-foreground font-mono">
                  {ch.group || '(全局)'}
                </TableCell>
                <TableCell className="text-sm text-right tabular-nums">{ch.priority}</TableCell>
                <TableCell className="text-sm text-right tabular-nums">{ch.weight}</TableCell>
                <TableCell>
                  {ch.status === 1 ? (
                    <Badge variant="default">启用</Badge>
                  ) : (
                    <Badge variant="outline">禁用</Badge>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
