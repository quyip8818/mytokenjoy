import { writeFileSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'
import { mockOperationLogs, mockCallLogs } from '../src/mocks/data'

const outDir = join(import.meta.dirname, '../../backend/internal/seed/data')
mkdirSync(outDir, { recursive: true })
writeFileSync(join(outDir, 'operation_logs.json'), JSON.stringify(mockOperationLogs, null, 2))
writeFileSync(join(outDir, 'call_logs.json'), JSON.stringify(mockCallLogs, null, 2))
console.log(
  `Exported ${mockOperationLogs.length} operation logs and ${mockCallLogs.length} call logs`,
)
