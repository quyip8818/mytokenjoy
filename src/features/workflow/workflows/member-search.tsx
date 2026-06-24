import { useState } from 'react'
import { Search } from 'lucide-react'
import type { Member } from '@/api/types'
import { memberApi } from '@/api/org'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'

export function MemberSearchWorkflow({
  entry,
  onPop,
  onClose,
  onSetDirty,
}: WorkflowComponentProps) {
  const excludeIds = (entry.payload.excludeIds as string[]) ?? []
  const onConfirm = entry.payload.onConfirm as ((members: Member[]) => void) | undefined
  const multi = entry.payload.multi !== false

  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<Member[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)

  const handleSearch = async () => {
    if (!keyword.trim()) return
    setLoading(true)
    try {
      const res = await memberApi.list({ page: 1, pageSize: 30, keyword: keyword.trim() })
      setResults(res.items.filter((m) => !excludeIds.includes(m.id)))
    } catch {
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  const toggleMember = (member: Member) => {
    if (!multi) {
      setSelected(new Set([member.id]))
      onSetDirty(true)
      return
    }
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(member.id)) next.delete(member.id)
      else next.add(member.id)
      return next
    })
    onSetDirty(true)
  }

  const handleConfirm = () => {
    const picked = results.filter((m) => selected.has(m.id))
    if (picked.length === 0) return
    onConfirm?.(picked)
    onPop()
  }

  return (
    <WorkflowPanelChrome
      title="搜索成员"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel="确认"
          onPrimary={handleConfirm}
          primaryDisabled={selected.size === 0}
        />
      }
    >
      <div className="space-y-4">
        <div className="flex gap-2">
          <Input
            placeholder="输入姓名搜索..."
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            className="flex-1"
          />
          <Button onClick={handleSearch} disabled={loading}>
            <Search className="h-4 w-4" />
          </Button>
        </div>
        <div className="max-h-[50vh] overflow-y-auto rounded-lg border border-border/60 divide-y divide-border/40">
          {results.length === 0 ? (
            <p className="px-4 py-8 text-center text-sm text-muted-foreground">
              {keyword ? '无匹配成员' : '请搜索成员'}
            </p>
          ) : (
            results.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => toggleMember(m)}
                className={cn(
                  'flex w-full items-center gap-3 px-4 py-3 text-left hover:bg-indigo-50/30',
                  selected.has(m.id) && 'bg-indigo-50/40',
                )}
              >
                {multi && (
                  <Checkbox checked={selected.has(m.id)} onCheckedChange={() => toggleMember(m)} />
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">{m.name}</p>
                  <p className="text-xs text-muted-foreground truncate">{m.departmentName}</p>
                </div>
              </button>
            ))
          )}
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
