import { readFileSync, existsSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const frontendRoot = join(scriptDir, '..')
const srcRoot = join(frontendRoot, 'src')

function fail(message) {
  console.error(`check-conventions: ${message}`)
  process.exit(1)
}

function readSrc(relativePath) {
  return readFileSync(join(srcRoot, relativePath), 'utf8')
}

function walkFiles(dir, callback) {
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

function relativeToSrc(absolutePath) {
  return absolutePath.slice(srcRoot.length + 1)
}

const routesSource = readSrc('config/routes.ts')

const definitionKeys = [...routesSource.matchAll(/key: '(\w+)'/g)].map((match) => match[1])
const definitionPaths = [...routesSource.matchAll(/path: '([^']+)'/g)]
  .map((match) => match[1])
  .filter((path) => path !== '/')

const lazyImports = [...routesSource.matchAll(/lazy: \(\) => import\('(@\/routes\/[^']+)'\)/g)].map(
  (match) => match[1],
)

if (definitionKeys.length !== lazyImports.length) {
  fail(
    `ROUTE_DEFINITIONS key count (${definitionKeys.length}) must match lazy import count (${lazyImports.length})`,
  )
}

const uniqueKeys = new Set(definitionKeys)
if (uniqueKeys.size !== definitionKeys.length) {
  fail('ROUTE_DEFINITIONS contains duplicate keys')
}

const uniquePaths = new Set(definitionPaths)
if (uniquePaths.size !== definitionPaths.length) {
  fail('ROUTE_DEFINITIONS contains duplicate paths')
}

for (const definition of ROUTE_DEFINITIONSFromSource(routesSource)) {
  if (!definition.navGroup) {
    fail(`ROUTE_DEFINITIONS entry ${definition.key} is missing navGroup`)
  }
}

const navPaths = [...routesSource.matchAll(/navGroup: '([^']+)'/g)]
if (navPaths.length !== definitionKeys.length) {
  fail('Every ROUTE_DEFINITIONS entry must declare navGroup')
}

for (const importPath of lazyImports) {
  const relativePath = importPath.replace('@/', '') + '.tsx'
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

function ROUTE_DEFINITIONSFromSource(source) {
  const entries = []
  const blocks = source.match(/\{\s*key: '[^']+'[\s\S]*?navGroup: '[^']+'[\s\S]*?\},/g) ?? []
  for (const block of blocks) {
    const key = block.match(/key: '(\w+)'/)?.[1]
    const navGroup = block.match(/navGroup: '([^']+)'/)?.[1]
    if (key) {
      entries.push({ key, navGroup })
    }
  }
  return entries
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
