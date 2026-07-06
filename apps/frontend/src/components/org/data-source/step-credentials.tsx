import { useState } from 'react'
import { useForm } from 'react-hook-form'
import type { Credential, Platform } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { CheckCircle2, Loader2, XCircle } from 'lucide-react'

interface StepCredentialsProps {
  platform: Platform
  onConnected: () => void
  onBack: () => void
}

interface FeishuFields {
  appId: string
  appSecret: string
}

interface DingtalkFields {
  corpId: string
  appKey: string
  appSecret: string
}

interface WecomFields {
  corpId: string
  secret: string
  agentId: string
}

type FormFields = FeishuFields & DingtalkFields & WecomFields

const platformLabels: Record<Platform, string> = {
  feishu: '飞书',
  dingtalk: '钉钉',
  wecom: '企业微信',
}

export function StepCredentials({ platform, onConnected, onBack }: StepCredentialsProps) {
  const apis = useInjectedApis()
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null)
  const [testMessage, setTestMessage] = useState('')

  const {
    register,
    getValues,
    formState: { errors },
    trigger,
  } = useForm<FormFields>()

  const buildCredential = (data: FormFields): Credential => {
    switch (platform) {
      case 'feishu':
        return { platform: 'feishu', appId: data.appId, appSecret: data.appSecret }
      case 'dingtalk':
        return { platform: 'dingtalk', corpId: data.corpId, appKey: data.appKey, appSecret: data.appSecret }
      case 'wecom':
        return { platform: 'wecom', corpId: data.corpId, secret: data.secret, agentId: data.agentId }
    }
  }

  const handleTest = async () => {
    const valid = await trigger()
    if (!valid) return

    setTesting(true)
    setTestResult(null)
    setTestMessage('')
    try {
      const data = getValues()
      const credential = buildCredential(data)
      const res = await apis.dataSourceApi.testConnection(credential)
      if (res.success) {
        setTestResult('success')
        setTestMessage('连接成功')
        await apis.dataSourceApi.save(credential)
      } else {
        setTestResult('error')
        setTestMessage(res.message || '连接失败，请检查凭证信息')
      }
    } catch {
      setTestResult('error')
      setTestMessage('连接测试出错，请稍后重试')
    } finally {
      setTesting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-semibold">配置 {platformLabels[platform]} 凭证</h3>
        <p className="text-sm text-muted-foreground mt-1">
          填写应用凭证信息并测试连接，确认能正常访问组织数据
        </p>
      </div>

      <div className="space-y-4 max-w-md">
        {platform === 'feishu' && (
          <>
            <div className="space-y-1.5">
              <Label htmlFor="appId">App ID</Label>
              <Input
                id="appId"
                placeholder="cli_xxxxxxxxxxxxxxxx"
                {...register('appId', { required: '请输入 App ID' })}
              />
              {errors.appId && (
                <p className="text-destructive text-xs">{errors.appId.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="appSecret">App Secret</Label>
              <Input
                id="appSecret"
                type="password"
                placeholder="输入 App Secret"
                {...register('appSecret', { required: '请输入 App Secret' })}
              />
              {errors.appSecret && (
                <p className="text-destructive text-xs">{errors.appSecret.message}</p>
              )}
            </div>
          </>
        )}
        {platform === 'dingtalk' && (
          <>
            {/* Permission reminder */}
            <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200">
              <p className="font-medium mb-1">请确认已开通以下权限</p>
              <ul className="list-disc pl-4 text-xs space-y-0.5 text-amber-700 dark:text-amber-300">
                <li>通讯录部门信息读权限</li>
                <li>成员信息读权限</li>
                <li>通讯录部门成员读权限</li>
              </ul>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="corpId">Corp ID</Label>
              <Input
                id="corpId"
                placeholder="输入企业 CorpID"
                {...register('corpId', { required: '请输入 CorpID' })}
              />
              {errors.corpId && (
                <p className="text-destructive text-xs">{errors.corpId.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="appKey">App Key</Label>
              <Input
                id="appKey"
                placeholder="输入应用的 App Key"
                {...register('appKey', { required: '请输入 App Key' })}
              />
              {errors.appKey && (
                <p className="text-destructive text-xs">{errors.appKey.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="appSecret">App Secret</Label>
              <Input
                id="appSecret"
                type="password"
                placeholder="输入应用的 App Secret"
                {...register('appSecret', { required: '请输入 Client Secret' })}
              />
              {errors.appSecret && (
                <p className="text-destructive text-xs">{errors.appSecret.message}</p>
              )}
            </div>
          </>
        )}
        {platform === 'wecom' && (
          <>
            <div className="space-y-1.5">
              <Label htmlFor="corpId">Corp ID</Label>
              <Input
                id="corpId"
                placeholder="输入企业 CorpID"
                {...register('corpId', { required: '请输入 CorpID' })}
              />
              {errors.corpId && (
                <p className="text-destructive text-xs">{errors.corpId.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="secret">Secret</Label>
              <Input
                id="secret"
                type="password"
                placeholder="输入 Secret"
                {...register('secret', { required: '请输入 Secret' })}
              />
              {errors.secret && (
                <p className="text-destructive text-xs">{errors.secret.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="agentId">Agent ID</Label>
              <Input
                id="agentId"
                placeholder="输入 AgentID"
                {...register('agentId', { required: '请输入 AgentID' })}
              />
              {errors.agentId && (
                <p className="text-destructive text-xs">{errors.agentId.message}</p>
              )}
            </div>
          </>
        )}
      </div>

      {/* Test result feedback */}
      {testResult && (
        <div
          className={`flex items-center gap-2 rounded-md border px-4 py-3 text-sm ${
            testResult === 'success'
              ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
              : 'border-red-200 bg-red-50 text-red-700'
          }`}
        >
          {testResult === 'success' ? (
            <CheckCircle2 className="size-4 shrink-0" />
          ) : (
            <XCircle className="size-4 shrink-0" />
          )}
          {testMessage}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-3 pt-2">
        <Button variant="outline" onClick={onBack}>
          上一步
        </Button>
        {testResult === 'success' ? (
          <Button onClick={onConnected}>下一步</Button>
        ) : (
          <Button onClick={handleTest} disabled={testing}>
            {testing && <Loader2 className="size-4 animate-spin" />}
            {testing ? '测试中...' : '测试连接'}
          </Button>
        )}
      </div>
    </div>
  )
}
