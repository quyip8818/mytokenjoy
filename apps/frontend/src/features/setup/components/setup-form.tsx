import { useCallback, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/ui/password-input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { SetupInitInput } from '@/api/setup'

interface SetupFormProps {
  submitting: boolean
  error: string | null
  onSubmit: (input: SetupInitInput) => void
}

export function SetupForm({ submitting, error, onSubmit }: SetupFormProps) {
  const [companyName, setCompanyName] = useState('')
  const [industry, setIndustry] = useState('')
  const [size, setSize] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [adminPassword, setAdminPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [adminName, setAdminName] = useState('')
  const [localError, setLocalError] = useState<string | null>(null)

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      setLocalError(null)
      if (adminPassword.length < 8) {
        setLocalError('密码至少需要 8 位')
        return
      }
      if (adminPassword !== confirmPassword) {
        setLocalError('两次密码输入不一致')
        return
      }
      onSubmit({
        companyName: companyName.trim(),
        industry,
        size,
        adminEmail: adminEmail.trim(),
        adminPassword,
        adminName: adminName.trim(),
      })
    },
    [companyName, industry, size, adminEmail, adminPassword, confirmPassword, adminName, onSubmit],
  )

  const displayError = error || localError

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5">
      {/* Company Name */}
      <div className="space-y-2">
        <Label htmlFor="company-name" className="text-sm font-medium">
          公司名称
        </Label>
        <Input
          id="company-name"
          type="text"
          placeholder="您的企业名称"
          className="h-11"
          value={companyName}
          onChange={(e) => setCompanyName(e.target.value)}
          required
        />
      </div>

      {/* Industry */}
      <div className="space-y-2">
        <Label className="text-sm font-medium">所属行业</Label>
        <Select value={industry} onValueChange={setIndustry}>
          <SelectTrigger className="!h-11 w-full">
            <SelectValue placeholder="请选择行业" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="互联网/科技">互联网/科技</SelectItem>
            <SelectItem value="金融">金融</SelectItem>
            <SelectItem value="教育">教育</SelectItem>
            <SelectItem value="医疗健康">医疗健康</SelectItem>
            <SelectItem value="电商/零售">电商/零售</SelectItem>
            <SelectItem value="制造业">制造业</SelectItem>
            <SelectItem value="游戏/娱乐">游戏/娱乐</SelectItem>
            <SelectItem value="企业服务">企业服务</SelectItem>
            <SelectItem value="政府/公共事业">政府/公共事业</SelectItem>
            <SelectItem value="其他">其他</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Size */}
      <div className="space-y-2">
        <Label className="text-sm font-medium">人员规模</Label>
        <Select value={size} onValueChange={setSize}>
          <SelectTrigger className="!h-11 w-full">
            <SelectValue placeholder="请选择人员规模" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1-10">1-10 人</SelectItem>
            <SelectItem value="11-50">11-50 人</SelectItem>
            <SelectItem value="51-200">51-200 人</SelectItem>
            <SelectItem value="201-500">201-500 人</SelectItem>
            <SelectItem value="501-1000">501-1000 人</SelectItem>
            <SelectItem value="1000+">1000 人以上</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Admin Email */}
      <div className="space-y-2">
        <Label htmlFor="admin-email" className="text-sm font-medium">
          管理员邮箱
        </Label>
        <Input
          id="admin-email"
          type="email"
          placeholder="admin@company.com"
          className="h-11"
          value={adminEmail}
          onChange={(e) => setAdminEmail(e.target.value)}
          required
        />
      </div>

      {/* Admin Name */}
      <div className="space-y-2">
        <Label htmlFor="admin-name" className="text-sm font-medium">
          管理员姓名
        </Label>
        <Input
          id="admin-name"
          type="text"
          placeholder="您的姓名"
          className="h-11"
          value={adminName}
          onChange={(e) => setAdminName(e.target.value)}
        />
      </div>

      {/* Password */}
      <div className="space-y-2">
        <Label htmlFor="admin-password" className="text-sm font-medium">
          设置密码
        </Label>
        <PasswordInput
          id="admin-password"
          autoComplete="new-password"
          placeholder="至少 8 位"
          className="h-11"
          value={adminPassword}
          onChange={(e) => setAdminPassword(e.target.value)}
          required
          minLength={8}
        />
      </div>

      {/* Confirm Password */}
      <div className="space-y-2">
        <Label htmlFor="confirm-password" className="text-sm font-medium">
          确认密码
        </Label>
        <PasswordInput
          id="confirm-password"
          autoComplete="new-password"
          placeholder="再次输入密码"
          className="h-11"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          required
          minLength={8}
        />
      </div>

      {/* Error */}
      {displayError && (
        <div
          className="rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          role="alert"
        >
          {displayError}
        </div>
      )}

      {/* Submit */}
      <Button
        type="submit"
        className="h-11 w-full text-base font-medium"
        disabled={submitting || !companyName.trim() || !adminEmail.trim()}
        disabledReason={submitting ? '初始化中…' : undefined}
      >
        {submitting ? '初始化中…' : '完成初始化'}
      </Button>
    </form>
  )
}
