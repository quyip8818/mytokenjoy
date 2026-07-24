import { MyCallLogsPageShell, useMyCallLogsPage } from '@/features/mydashboard'

export default function MyUsagePage() {
  return <MyCallLogsPageShell {...useMyCallLogsPage()} />
}
