import { useState } from 'react'
import { useForm } from 'react-hook-form'
import type { Credential, Platform } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { PLATFORM_LABELS } from '@/features/org'

interface CredentialFormProps {
  connected: boolean
  currentPlatform: Platform | null
  dataSourceApi: AppApis['dataSourceApi']
  onSaved: () => void
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

export function CredentialForm({
  connected,
  currentPlatform,
  dataSourceApi,
  onSaved,
}: CredentialFormProps) {
  const [platform, setPlatform] = useState<Platform>(currentPlatform || 'feishu')
  const [testSuccess, setTestSuccess] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testMessage, setTestMessage] = useState('')
  const [saving, setSaving] = useState(false)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchResult, setSearchResult] = useState<{
    name: string
    department: string
    mappingOk: boolean
  } | null>(null)
  const [showPlatformConfirm, setShowPlatformConfirm] = useState(false)
  const [showSaveConfirm, setShowSaveConfirm] = useState(false)
  const [pendingPlatform, setPendingPlatform] = useState<Platform | null>(null)

  const {
    register,
    handleSubmit,
    getValues,
    reset,
    formState: { errors },
  } = useForm<FormFields>()

  const handlePlatformChange = (newPlatform: Platform) => {
    if (newPlatform === platform) return
    if (connected || testSuccess) {
      setPendingPlatform(newPlatform)
      setShowPlatformConfirm(true)
    } else {
      switchPlatform(newPlatform)
    }
  }

  const switchPlatform = (newPlatform: Platform) => {
    setPlatform(newPlatform)
    setTestSuccess(false)
    setTestMessage('')
    setSearchResult(null)
    reset()
  }

  const buildCredential = (data: FormFields): Credential => {
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

  const handleTestConnection = async () => {
    const data = getValues()
    const credential = buildCredential(data)
    setTesting(true)
    setTestMessage('')
    try {
      const res = await dataSourceApi.testConnection(credential)
      if (res.success) {
        setTestSuccess(true)
        setTestMessage('连接成功')
      } else {
        setTestMessage(res.message || '连接失败')
      }
    } catch {
      setTestMessage('连接测试出错')
    } finally {
      setTesting(false)
    }
  }

  const handleSearch = async () => {
    if (!searchKeyword.trim()) return
    const result = await dataSourceApi.searchMember(searchKeyword)
    setSearchResult(result)
  }

  const onSubmit = (data: FormFields) => {
    if (connected) {
      setShowSaveConfirm(true)
    } else {
      doSave(data)
    }
  }

  const doSave = async (data?: FormFields) => {
    const values = data || getValues()
    const credential = buildCredential(values)
    setSaving(true)
    try {
      await dataSourceApi.save(credential)
      onSaved()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      {/* Platform Radio */}
      <div className="mb-6">
        <Label className="mb-2">选择平台</Label>
        <RadioGroup
          value={platform}
          onValueChange={(value) => handlePlatformChange(value as Platform)}
          className="flex flex-row gap-4"
        >
          {(['feishu', 'dingtalk', 'wecom'] as Platform[]).map((p) => (
            <Label key={p} className="cursor-pointer">
              <RadioGroupItem value={p} />
              <span className="text-sm">{PLATFORM_LABELS[p]}</span>
            </Label>
          ))}
        </RadioGroup>
      </div>

      {/* Credential Fields */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {platform === 'feishu' && (
          <>
            <div>
              <Label className="mb-1">App ID</Label>
              <Input {...register('appId', { required: '请输入 App ID' })} />
              {errors.appId && (
                <p className="text-destructive text-xs mt-1">{errors.appId.message}</p>
              )}
            </div>
            <div>
              <Label className="mb-1">App Secret</Label>
              <Input
                type="password"
                {...register('appSecret', { required: '请输入 App Secret' })}
              />
              {errors.appSecret && (
                <p className="text-destructive text-xs mt-1">{errors.appSecret.message}</p>
              )}
            </div>
          </>
        )}
        {platform === 'dingtalk' && (
          <>
            <div>
              <Label className="mb-1">CorpID</Label>
              <Input {...register('corpId', { required: '请输入 CorpID' })} />
              {errors.corpId && (
                <p className="text-destructive text-xs mt-1">{errors.corpId.message}</p>
              )}
            </div>
            <div>
              <Label className="mb-1">AppKey</Label>
              <Input {...register('appKey', { required: '请输入 AppKey' })} />
              {errors.appKey && (
                <p className="text-destructive text-xs mt-1">{errors.appKey.message}</p>
              )}
            </div>
            <div>
              <Label className="mb-1">AppSecret</Label>
              <Input type="password" {...register('appSecret', { required: '请输入 AppSecret' })} />
              {errors.appSecret && (
                <p className="text-destructive text-xs mt-1">{errors.appSecret.message}</p>
              )}
            </div>
          </>
        )}
        {platform === 'wecom' && (
          <>
            <div>
              <Label className="mb-1">CorpID</Label>
              <Input {...register('corpId', { required: '请输入 CorpID' })} />
              {errors.corpId && (
                <p className="text-destructive text-xs mt-1">{errors.corpId.message}</p>
              )}
            </div>
            <div>
              <Label className="mb-1">Secret</Label>
              <Input type="password" {...register('secret', { required: '请输入 Secret' })} />
              {errors.secret && (
                <p className="text-destructive text-xs mt-1">{errors.secret.message}</p>
              )}
            </div>
            <div>
              <Label className="mb-1">AgentID</Label>
              <Input {...register('agentId', { required: '请输入 AgentID' })} />
              {errors.agentId && (
                <p className="text-destructive text-xs mt-1">{errors.agentId.message}</p>
              )}
            </div>
          </>
        )}

        {/* Test Connection */}
        <div className="flex items-center gap-3">
          <Button type="button" variant="outline" onClick={handleTestConnection} disabled={testing}>
            {testing ? '测试中...' : '测试连接'}
          </Button>
          {testMessage && (
            <span className={`text-sm ${testSuccess ? 'text-green-600' : 'text-destructive'}`}>
              {testMessage}
            </span>
          )}
        </div>

        {/* Search test area */}
        {testSuccess && (
          <div className="rounded-md border border-border bg-muted/50 p-4">
            <p className="text-sm font-medium text-foreground mb-2">搜索成员验证映射</p>
            <div className="flex gap-2">
              <Input
                value={searchKeyword}
                onChange={(e) => setSearchKeyword(e.target.value)}
                placeholder="输入姓名搜索"
                className="flex-1"
              />
              <Button type="button" variant="outline" onClick={handleSearch}>
                搜索
              </Button>
            </div>
            {searchResult && (
              <div className="mt-3 text-sm">
                <p>姓名：{searchResult.name}</p>
                <p>部门：{searchResult.department}</p>
                <p>
                  映射状态：
                  {searchResult.mappingOk ? (
                    <span className="text-green-600">正常</span>
                  ) : (
                    <span className="text-destructive">异常</span>
                  )}
                </p>
              </div>
            )}
          </div>
        )}

        {/* Save */}
        <Button type="submit" disabled={!testSuccess || saving}>
          {saving ? '保存中...' : '保存凭证'}
        </Button>
      </form>

      {/* Platform switch confirm */}
      <AlertDialog
        open={showPlatformConfirm}
        onOpenChange={(open) => {
          if (!open) {
            setShowPlatformConfirm(false)
            setPendingPlatform(null)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>切换平台</AlertDialogTitle>
            <AlertDialogDescription>切换平台将清空当前配置，确定要切换吗？</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setShowPlatformConfirm(false)
                setPendingPlatform(null)
              }}
            >
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                setShowPlatformConfirm(false)
                if (pendingPlatform) switchPlatform(pendingPlatform)
                setPendingPlatform(null)
              }}
            >
              确定
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Save confirm for existing credential */}
      <AlertDialog
        open={showSaveConfirm}
        onOpenChange={(open) => {
          if (!open) setShowSaveConfirm(false)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>修改凭证</AlertDialogTitle>
            <AlertDialogDescription>
              修改凭证后需要重新导入数据，确定保存吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setShowSaveConfirm(false)}>取消</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                setShowSaveConfirm(false)
                doSave()
              }}
            >
              确定
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
