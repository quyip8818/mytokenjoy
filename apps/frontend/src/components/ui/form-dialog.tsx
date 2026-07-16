'use client'

import * as React from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface FormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: React.ReactNode
  error?: string | null
  busy?: boolean
  submitLabel?: string
  submitDisabled?: boolean
  cancelLabel?: string
  onSubmit: () => void | Promise<void>
  className?: string
  children: React.ReactNode
}

function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  error,
  busy = false,
  submitLabel = '提交',
  submitDisabled = false,
  cancelLabel = '取消',
  onSubmit,
  className,
  children,
}: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-md', className)}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && (
            <DialogDescription asChild={typeof description !== 'string'}>
              {typeof description === 'string' ? description : description}
            </DialogDescription>
          )}
        </DialogHeader>

        <div className="space-y-4">{children}</div>

        {error && (
          <pre className="bg-destructive/10 text-destructive max-h-40 overflow-auto rounded-md p-3 text-xs whitespace-pre-wrap break-all">
            {error}
          </pre>
        )}

        <DialogFooter>
          <Button variant="outline" size="sm" onClick={() => onOpenChange(false)} disabled={busy}>
            {cancelLabel}
          </Button>
          <Button size="sm" onClick={() => void onSubmit()} disabled={busy || submitDisabled}>
            {busy ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                {submitLabel}中…
              </>
            ) : (
              submitLabel
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export { FormDialog }
export type { FormDialogProps }
