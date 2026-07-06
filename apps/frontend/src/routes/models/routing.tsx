import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { routingApi } from '@/api/models'
import type { RoutingRule } from '@/api/types'

export default function ModelRoutingPage() {
  const [rules, setRules] = useState<RoutingRule[]>([])

  useEffect(() => {
    routingApi.getRules().then(setRules)
  }, [])

  return (
    <div className="space-y-8">
      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">组织节点模型白名单</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">组织节点</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">可用模型</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">默认模型</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">降级模型</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">来源</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rules.map((rule) => (
                <TableRow key={rule.id} className="border-border-subtle hover:bg-muted/50">
                  <TableCell className="font-medium">{rule.nodeName}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {rule.allowedModels.map(m => (
                        <Badge key={m} variant="outline" className="bg-primary/10 text-primary border-primary/20 text-[11px]">{m}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell>
                    {rule.defaultModel ? <Badge variant="secondary" className="bg-violet-50 text-violet-700 border-0">{rule.defaultModel}</Badge> : <span className="text-muted-foreground text-sm">-</span>}
                  </TableCell>
                  <TableCell>
                    {rule.fallbackModel ? <Badge variant="secondary" className="bg-violet-50 text-violet-700 border-0">{rule.fallbackModel}</Badge> : <span className="text-muted-foreground text-sm">-</span>}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={`border-0 ${rule.inherited ? 'bg-slate-100 text-slate-600' : 'bg-primary/10 text-primary'}`}>
                      {rule.inherited ? '继承' : '自定义'}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
