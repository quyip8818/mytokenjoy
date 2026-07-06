import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { providerKeyApi } from '@/api/keys'
import type { ProviderKey } from '@/api/types'

const providerLabels: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  qwen: '通义千问',
  custom: '自定义',
}

const providerBadgeStyles: Record<string, string> = {
  openai: 'bg-emerald-50 text-emerald-700 border-emerald-100',
  anthropic: 'bg-orange-50 text-orange-700 border-orange-100',
  deepseek: 'bg-blue-50 text-blue-700 border-blue-100',
  qwen: 'bg-purple-50 text-purple-700 border-purple-100',
  custom: 'bg-slate-50 text-slate-700 border-slate-100',
}

export default function ProviderKeysPage() {
  const [keys, setKeys] = useState<ProviderKey[]>([])
  const [formOpen, setFormOpen] = useState(false)
  const [provider, setProvider] = useState('openai')
  const [name, setName] = useState('')
  const [keyValue, setKeyValue] = useState('')

  useEffect(() => {
    providerKeyApi.list().then(setKeys)
  }, [])

  const handleCreate = async () => {
    if (!name || !keyValue) return
    const key = await providerKeyApi.create({ provider, name, key: keyValue })
    setKeys([...keys, key])
    setFormOpen(false)
    setName('')
    setKeyValue('')
  }

  const handleToggle = async (key: ProviderKey) => {
    const enabled = key.status !== 'active'
    await providerKeyApi.toggle(key.id, enabled)
    setKeys(keys.map(k => k.id === key.id ? { ...k, status: enabled ? 'active' : 'disabled' } : k))
  }

  const handleDelete = async (id: string) => {
    await providerKeyApi.delete(id)
    setKeys(keys.filter(k => k.id !== id))
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700">正常</span>
      case 'disabled':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600">已禁用</span>
      case 'error':
      case 'expired':
        return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700">{status === 'error' ? '异常' : '已过期'}</span>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getBalanceClass = (balance: number | null) => {
    if (balance === null) return 'text-muted-foreground'
    if (balance > 1000) return 'text-emerald-600 font-medium'
    if (balance < 500) return 'text-amber-600 font-medium'
    return ''
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-end">
        <Button
          size="sm"
          onClick={() => setFormOpen(true)}
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          添加 Key
        </Button>
      </div>

      <Card className="shadow-xs border-border">
        <CardContent className="pt-5 pb-4">
          <h3 className="text-sm font-semibold text-foreground/80 mb-4">Key 池</h3>
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">名称</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">供应商</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">Key 前缀</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">状态</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">余额</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">最后使用</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((key) => (
                <TableRow key={key.id} className="border-border-subtle hover:bg-muted/50">
                  <TableCell className="font-medium">{key.name}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className={providerBadgeStyles[key.provider] ?? providerBadgeStyles.custom}>
                      {providerLabels[key.provider]}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className="font-mono text-xs px-1.5 py-0.5 rounded bg-indigo-50 text-muted-foreground">{key.keyPrefix}</span>
                  </TableCell>
                  <TableCell>{getStatusBadge(key.status)}</TableCell>
                  <TableCell className={`text-right ${getBalanceClass(key.balance)}`}>
                    {key.balance !== null ? `$${key.balance.toFixed(2)}` : '-'}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{key.lastUsed ?? '-'}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="sm" onClick={() => handleToggle(key)}>
                        {key.status === 'active' ? '禁用' : '启用'}
                      </Button>
                      <Button variant="ghost" size="sm" className="text-destructive" onClick={() => handleDelete(key.id)}>删除</Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={formOpen} onOpenChange={(o) => { if (!o) setFormOpen(false) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>添加供应商 Key</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>供应商</Label>
              <Select value={provider} onValueChange={(v) => setProvider(v ?? 'openai')}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="anthropic">Anthropic</SelectItem>
                  <SelectItem value="deepseek">DeepSeek</SelectItem>
                  <SelectItem value="qwen">通义千问</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>名称</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="例如：OpenAI 主力" />
            </div>
            <div className="space-y-2">
              <Label>API Key</Label>
              <Input value={keyValue} onChange={(e) => setKeyValue(e.target.value)} placeholder="sk-..." type="password" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setFormOpen(false)}>取消</Button>
            <Button
              onClick={handleCreate}
              className="bg-primary text-primary-foreground hover:bg-primary/90"
            >
              添加
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
