import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { getRouteLazyImportPaths, validateRouteDefinitions } from '../src/config/routes.ts'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const frontendRoot = join(scriptDir, '..')
const srcRoot = join(frontendRoot, 'src')

function fail(message: string): never {
  console.error(`check-conventions: ${message}`)
  process.exit(1)
}

function walkFiles(dir: string, callback: (filePath: string) => void) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const fullPath = join(dir, entry.name)
    if (entry.isDirectory()) {
      walkFiles(fullPath, callback)
      continue
    }
    if (/\.(ts|tsx)$/.test(entry.name)) {
      callback(fullPath)
    }
  }
}

function relativeToSrc(absolutePath: string) {
  return absolutePath.slice(srcRoot.length + 1)
}

try {
  validateRouteDefinitions()
} catch (error) {
  fail(error instanceof Error ? error.message : String(error))
}

for (const importPath of getRouteLazyImportPaths()) {
  const relativePath = `${importPath.replace('@/', '')}.tsx`
  const pagePath = join(srcRoot, relativePath)
  if (!existsSync(pagePath)) {
    fail(`ROUTE_DEFINITIONS lazy import target not found: ${relativePath}`)
  }

  const pageSource = readFileSync(pagePath, 'utf8')
  const hasPageHookImport = /from ['"]\.\/hooks\/|from ['"]@\/routes\/[^'"]+\/hooks\//.test(
    pageSource,
  )
  if (!hasPageHookImport) {
    fail(`Page ${relativePath} must import a hook from ./hooks/ (use-*-page.ts pattern)`)
  }
}

const uiDir = join(srcRoot, 'components/ui')
const domainNamePattern =
  /budget|org|audit|keys|model|member|department|role|sync|credential|approval/i

for (const file of readdirSync(uiDir)) {
  if (!file.endsWith('.tsx')) continue
  if (domainNamePattern.test(file)) {
    fail(
      `components/ui/${file} looks domain-specific; use components/{domain}/ or routes/{domain}/components/`,
    )
  }
}

const componentsDir = join(srcRoot, 'components')
walkFiles(componentsDir, (filePath) => {
  const source = readFileSync(filePath, 'utf8')
  for (const line of source.split('\n')) {
    const trimmed = line.trimStart()
    if (
      (trimmed.includes("from '@/routes/") || trimmed.includes('from "@/routes/')) &&
      trimmed.startsWith('import') &&
      !trimmed.startsWith('import type')
    ) {
      fail(
        `${relativeToSrc(filePath)}: components/ must not import from @/routes/ (use hooks/ or lib/)`,
      )
    }
  }
})

walkFiles(join(srcRoot, 'routes'), (filePath) => {
  if (!filePath.includes('/components/')) return
  const source = readFileSync(filePath, 'utf8')
  if (/\buseApis\s*\(/.test(source)) {
    fail(
      `${relativeToSrc(filePath)}: routes/*/components/ must not call useApis(); lift data fetching to use-*-page.ts`,
    )
  }
})

console.log('check-conventions: all checks passed')
