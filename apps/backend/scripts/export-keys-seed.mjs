import { writeFileSync, mkdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { mockPlatformKeys } from '../../frontend/src/mocks/fixtures/keys.ts'

const __dirname = dirname(fileURLToPath(import.meta.url))
const outDir = join(__dirname, '../internal/seed/data')
mkdirSync(outDir, { recursive: true })
writeFileSync(join(outDir, 'platform_keys.json'), JSON.stringify(mockPlatformKeys, null, 2))
console.log(`Exported ${mockPlatformKeys.length} platform keys`)
