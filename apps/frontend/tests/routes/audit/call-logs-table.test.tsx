import { useState } from 'react'
import { fireEvent, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { CallLog } from '@/api/types'
import { CallLogsTable } from '@/routes/audit/components/call-logs-table'
import { renderWithProviders } from '@tests/utils'

const baseLog: CallLog = {
  id: 'cl-1',
  caller: '张三',
  callerId: 'm-1',
  callerType: 'member',
  model: 'gpt-4o',
  provider: 'openai',
  inputTokens: 100,
  outputTokens: 50,
  latencyMs: 200,
  status: 'success',
  cost: 1.5,
  createdAt: '2026-06-19 10:00',
  previewSnippet: 'hello world',
}

function TestHarness({
  logs,
  contentRetentionEnabled,
}: {
  logs: CallLog[]
  contentRetentionEnabled: boolean
}) {
  const [expandedId, setExpandedId] = useState<string | null>(null)

  return (
    <CallLogsTable
      logs={logs}
      expandedId={expandedId}
      contentRetentionEnabled={contentRetentionEnabled}
      onToggleExpanded={(id) => setExpandedId((current) => (current === id ? null : id))}
    />
  )
}

describe('CallLogsTable', () => {
  it('does not expand rows when content retention is disabled', () => {
    renderWithProviders(<TestHarness logs={[baseLog]} contentRetentionEnabled={false} />)

    expect(screen.queryByText('输入预览')).not.toBeInTheDocument()
    fireEvent.click(screen.getByText('张三'))
    expect(screen.queryByText('hello world')).not.toBeInTheDocument()
  })

  it('expands and shows preview snippet when content retention is enabled', () => {
    renderWithProviders(<TestHarness logs={[baseLog]} contentRetentionEnabled={true} />)

    fireEvent.click(screen.getByText('张三'))
    expect(screen.getByText('输入预览')).toBeInTheDocument()
    expect(screen.getByText('hello world')).toBeInTheDocument()
  })

  it('shows placeholder when preview snippet is empty', () => {
    renderWithProviders(
      <TestHarness logs={[{ ...baseLog, previewSnippet: '' }]} contentRetentionEnabled={true} />,
    )

    fireEvent.click(screen.getByText('张三'))
    expect(screen.getByText('—')).toBeInTheDocument()
  })
})
