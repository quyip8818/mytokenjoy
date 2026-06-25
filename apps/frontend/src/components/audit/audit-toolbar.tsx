import { Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { useAuditSettings } from '@/routes/audit/hooks/use-audit-settings'

interface AuditToolbarProps {
  onExport: () => void
  exportLabel?: string
}

export function AuditToolbar({ onExport, exportLabel = '导出 CSV' }: AuditToolbarProps) {
  const { contentRetentionEnabled, updateContentRetention, loading } = useAuditSettings()

  return (
    <div className="flex items-center gap-4">
      <div className="flex items-center gap-2">
        <Switch
          id="content-retention"
          checked={contentRetentionEnabled}
          disabled={loading}
          onCheckedChange={(checked) => void updateContentRetention(checked)}
        />
        <Label htmlFor="content-retention" className="text-sm text-muted-foreground">
          内容留存
        </Label>
      </div>
      <Button variant="outline" size="sm" onClick={onExport}>
        <Download className="mr-2 h-4 w-4" />
        {exportLabel}
      </Button>
    </div>
  )
}
