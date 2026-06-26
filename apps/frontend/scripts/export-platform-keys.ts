import { writeFileSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'
import { mockPlatformKeys } from '../src/mocks/data'

const outDir = join(import.meta.dirname, '../../backend/internal/seed/data')
mkdirSync(outDir, { recursive: true })
writeFileSync(join(outDir, 'platform_keys.json'), JSON.stringify(mockPlatformKeys, null, 2))
console.log(`Exported ${mockPlatformKeys.length} platform keys`)
