import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router'
import { budgetApi } from '@/api/budget'
import type { OverrunPolicyConfig } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoCta } from '@/features/demo'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

export default function BudgetAlertsPage() {
  const { open } = useWorkflow()
  const overrunCta = useDemoCta('OVERRUN')
  const [policy, setPolicy] = useState<OverrunPolicyConfig | null>(null)

  const load = useCallback(async () => {
    const p = await budgetApi.getOverrunPolicy()
    setPolicy(p)
  }, [])

  useEffect(() => {
    void budgetApi.getOverrunPolicy().then(setPolicy)
  }, [])

  const notifyLabels = [
    policy?.notifyEmail && '邮箱',
    policy?.notifyPhone && '手机',
    policy?.notifyIm && 'IM',
  ].filter(Boolean)

  return (
    <div className="space-y-6">
      <Card className="shadow-card border-border/50">
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle className="text-base">全局超限策略</CardTitle>
          <Button
            id={overrunCta.id}
            size="sm"
            className={cn(
              'bg-gradient-to-r from-indigo-600 to-violet-600 text-white',
              overrunCta.className,
            )}
            onClick={() => open('overrun-policy', { onSuccess: load })}
          >
            编辑策略
          </Button>
        </CardHeader>
        <CardContent className="space-y-4 text-sm">
          {policy ? (
            <>
              <div>
                <p className="text-muted-foreground mb-2">预警阈值</p>
                <div className="flex gap-2">
                  {policy.thresholds.map((t) => (
                    <Badge key={t} variant="outline">
                      {t}%
                    </Badge>
                  ))}
                </div>
              </div>
              <div>
                <p className="text-muted-foreground mb-1">通知渠道</p>
                <p>{notifyLabels.length > 0 ? notifyLabels.join('、') : '未配置'}</p>
              </div>
              <div>
                <p className="text-muted-foreground mb-1">超限阻断文案</p>
                <p className="text-foreground/90">{policy.blockMessage}</p>
              </div>
              <p className="text-xs text-muted-foreground">
                超限行为固定为直接阻断。组织预算分配见{' '}
                <Link to="/budget/overview" className="text-indigo-600 hover:underline">
                  预算总览
                </Link>
                。
              </p>
            </>
          ) : (
            <p className="text-muted-foreground">加载中...</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
