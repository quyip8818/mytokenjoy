import { useState } from 'react'
import { Upload, Download, FileText, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useApis } from '@/api/use-apis'
import { CONTRACT_STATUS } from '@/config/enums'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/badge'
import { ConfirmActionDialog, type ConfirmActionState } from '@/components/ui/confirm-action-dialog'
import { formatAmount } from '@/lib/utils'
import type { ContractDetail, ContractAttachment } from '@/api/contracts'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome } from '../components/workflow-panel-chrome'

export function ContractDetailWorkflow({
  entry,
  onClose,
}: WorkflowComponentProps<'contract-detail'>) {
  const apis = useApis()
  const { canEdit, onRefresh } = entry.payload
  const [detail, setDetail] = useState<ContractDetail>(entry.payload.contract)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const refreshDetail = async () => {
    const d = await apis.contractsApi.detail(detail.id)
    setDetail(d)
    onRefresh?.()
  }

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      await apis.contractsApi.uploadAttachment(detail.id, file)
      toast.success('上传成功')
      await refreshDetail()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '上传失败')
    }
    e.target.value = ''
  }

  const handleDeleteAttachment = (att: ContractAttachment) => {
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除附件「${att.fileName}」吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: async () => {
        await apis.contractsApi.deleteAttachment(detail.id, att.id)
        toast.success('删除成功')
        await refreshDetail()
        setConfirmState(null)
      },
    })
  }

  return (
    <WorkflowPanelChrome title={detail.title} onClose={onClose}>
      <div className="space-y-6">
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-muted-foreground">编号：</span>
            {detail.contractNo}
          </div>
          <div>
            <span className="text-muted-foreground">供应商：</span>
            {detail.supplierName}
          </div>
          <div>
            <span className="text-muted-foreground">金额：</span>
            {formatAmount(detail.amount)}
          </div>
          <div>
            <span className="text-muted-foreground">状态：</span>
            <StatusBadge status={detail.status} map={CONTRACT_STATUS} />
          </div>
          <div>
            <span className="text-muted-foreground">签订：</span>
            {detail.signDate ?? '-'}
          </div>
          <div>
            <span className="text-muted-foreground">生效：</span>
            {detail.startDate ?? '-'}
          </div>
          <div>
            <span className="text-muted-foreground">到期：</span>
            {detail.endDate ?? '-'}
          </div>
          <div>
            <span className="text-muted-foreground">备注：</span>
            {detail.remarks ?? '-'}
          </div>
        </div>

        <Card>
          <CardHeader className="flex-row items-center justify-between pb-3">
            <CardTitle className="text-sm">附件 ({detail.attachments.length})</CardTitle>
            {canEdit && (
              <Button size="sm" asChild>
                <label className="cursor-pointer">
                  <Upload className="h-3.5 w-3.5" /> 上传
                  <input type="file" className="hidden" onChange={handleUpload} />
                </label>
              </Button>
            )}
          </CardHeader>
          <CardContent>
            {detail.attachments.length === 0 ? (
              <div className="py-8 text-center text-sm text-muted-foreground">暂无附件</div>
            ) : (
              <div className="space-y-2">
                {detail.attachments.map((att) => (
                  <div
                    key={att.id}
                    className="flex items-center justify-between rounded border p-2.5"
                  >
                    <div className="flex min-w-0 items-center gap-2">
                      <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <div className="truncate text-sm">{att.fileName}</div>
                      <span className="text-xs text-muted-foreground">
                        {(att.fileSize / 1024).toFixed(0)} KB
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" asChild>
                        <a href={`/api/contracts/${detail.id}/attachments/${att.id}/download`}>
                          <Download className="h-3.5 w-3.5" />
                        </a>
                      </Button>
                      {canEdit && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7"
                          onClick={() => handleDeleteAttachment(att)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <ConfirmActionDialog
        state={confirmState}
        onOpenChange={(open) => !open && setConfirmState(null)}
        onClose={() => setConfirmState(null)}
      />
    </WorkflowPanelChrome>
  )
}
