import { Input } from '@/components/ui/input'

interface AuditKeywordInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

export function AuditKeywordInput({
  value,
  onChange,
  placeholder = '关键词搜索',
}: AuditKeywordInputProps) {
  return (
    <Input
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className="w-40 border-border/60"
    />
  )
}
