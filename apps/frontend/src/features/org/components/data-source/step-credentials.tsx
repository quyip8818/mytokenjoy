import { useState } from 'react'
import { useForm } from 'react-hook-form'
import type { Credential, Platform } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
import { ApiError } from '@/api/client'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { CheckCircle2, Loader2, XCircle } from 'lucide-react'
import { PLATFORM_LABELS } from '@/features/org'

interface StepCredentialsProps {
  platform: Platform
  dataSourceApi: AppApis['dataSourceApi']
  onConnected: () => void
  onBack: () => void
}

interface FieldConfig {
  name: keyof FormFields
  label: string
  placeholder: string
  password?: boolean
}

interface FormFields {
  appId: string
  appSecret: string
  corpId: string
  appKey: string
  secret: string
  agentId: string
}

const PLATFORM_FIELDS: Record<Platform, FieldConfig[]> = {
  feishu: [
    { name: 'appId', label: 'App ID', placeholder: 'cli_xxxxxxxxxxxxxxxx' },
    { name: 'appSecret', label: 'App Secret', placeholder: '输入 App Secret', password: true },
  ],
  dingtalk: [
    { name: 'corpId', label: 'Corp ID', placeholder: '输入企业 CorpID' },
    { name: 'appKey', label: 'App Key', placeholder: '输入应用的 App Key' },
    {
      name: 'appSecret',
      label: 'App Secret',
      placeholder: '输入应用的 App Secret',
      password: true,
    },
  ],
  wecom: [
    { name: 'corpId', label: 'Corp ID', placeholder: '输入企业 CorpID' },
    { name: 'secret', label: 'Secret', placeholder: '输入 Secret', password: true },
    { name: 'agentId', label: 'Agent ID', placeholder: '输入 AgentID' },
  ],
}

function buildCredential(platform: Platform, data: FormFields): Credential {
  switch (platform) {
    case 'feishu':
      return { platform: 'feishu', appId: data.appId, appSecret: data.appSecret }
    case 'dingtalk':
      return {
        platform: 'dingtalk',
        corpId: data.corpId,
        appKey: data.appKey,
        appSecret: data.appSecret,
      }
    case 'wecom':
      return {
        platform: 'wecom',
        corpId: data.corpId,
        secret: data.secret,
        agentId: data.agentId,
      }
  }
}

export function StepCredentials({
  platform,
  dataSourceApi,
  onConnected,
  onBack,
}: StepCredentialsProps) {
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null)
  const [testMessage, setTestMessage] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormFields>()

  // 凭证被修改后,之前的测试结果不再可信,回到待测试状态
  const handleFormChange = () => {
    if (testResult) {
      setTestResult(null)
      setTestMessage('')
    }
  }

  const handleTest = async (data: FormFields) => {
    setTesting(true)
    setTestResult(null)
    setTestMessage('')
    try {
      const credential = buildCredential(platform, data)
      const res = await dataSourceApi.testConnection(credential)
      if (res.success) {
        await dataSourceApi.save(credential)
        setTestResult('success')
        setTestMessage('连接成功，凭证已保存')
      } else {
        setTestResult('error')
        setTestMessage(res.message || '连接失败，请检查凭证信息')
      }
    } catch (err) {
      setTestResult('error')
      setTestMessage(
        err instanceof ApiError && err.message
          ? `连接失败：${err.message}`
          : '连接测试出错，请稍后重试',
      )
    } finally {
      setTesting(false)
    }
  }

  const fields = PLATFORM_FIELDS[platform]

  return (
    <form
      className="space-y-6"
      onChange={handleFormChange}
      onSubmit={(e) => {
        if (testResult === 'success') {
          e.preventDefault()
          onConnected()
          return
        }
        void handleSubmit(handleTest)(e)
      }}
    >
      <div>
        <h3 className="text-sm font-semibold">配置{PLATFORM_LABELS[platform]}凭证</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          填写应用凭证信息并测试连接，确认能正常访问组织数据
        </p>
      </div>

      <div className="max-w-md space-y-4">
        {platform === 'dingtalk' && (
          <div className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200">
            <p className="font-medium mb-1">请确认已开通以下权限</p>
            <ul className="list-disc pl-4 text-xs space-y-0.5 text-amber-700 dark:text-amber-300">
              <li>通讯录部门信息读权限</li>
              <li>成员信息读权限</li>
              <li>通讯录部门成员读权限</li>
            </ul>
          </div>
        )}

        {fields.map((field) => (
          <div key={field.name} className="space-y-1.5">
            <Label htmlFor={field.name}>{field.label}</Label>
            <Input
              id={field.name}
              type={field.password ? 'password' : 'text'}
              placeholder={field.placeholder}
              aria-invalid={Boolean(errors[field.name])}
              {...register(field.name, { required: `请输入 ${field.label}` })}
            />
            {errors[field.name] && (
              <p className="text-destructive text-xs">{errors[field.name]?.message}</p>
            )}
          </div>
        ))}
      </div>

      {testResult && (
        <div
          role="status"
          className={`flex max-w-md items-start gap-2 rounded-lg border px-4 py-3 text-sm ${
            testResult === 'success'
              ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
              : 'border-red-200 bg-red-50 text-red-700'
          }`}
        >
          {testResult === 'success' ? (
            <CheckCircle2 className="mt-0.5 size-4 shrink-0" />
          ) : (
            <XCircle className="mt-0.5 size-4 shrink-0" />
          )}
          <span className="break-all">{testMessage}</span>
        </div>
      )}

      <div className="flex items-center gap-3 border-t pt-4">
        <Button type="button" variant="outline" onClick={onBack}>
          上一步
        </Button>
        {testResult === 'success' ? (
          <Button type="submit">下一步</Button>
        ) : (
          <Button type="submit" disabled={testing}>
            {testing && <Loader2 className="size-4 animate-spin" />}
            {testing ? '测试中...' : '测试连接'}
          </Button>
        )}
      </div>
    </form>
  )
}
