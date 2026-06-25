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
const navSource = readSrc('config/nav.ts')

const routeMetaKeys = [...routesSource.matchAll(/path: (ROUTES\.\w+),\s*\n\s*label:/g)].map(
  (match) => match[1],
)

const appRouteKeys = [...routesSource.matchAll(/\{ path: (ROUTES\.\w+), lazy:/g)].map(
  (match) => match[1],
)

const navRouteKeys = [...navSource.matchAll(/ROUTES\.(\w+)/g)].map((match) => `ROUTES.${match[1]}`)

const uniqueMeta = new Set(routeMetaKeys)
const uniqueApp = new Set(appRouteKeys)
const uniqueNav = new Set(navRouteKeys)

if (uniqueMeta.size !== routeMetaKeys.length) {
  fail('ROUTE_META contains duplicate route keys')
}

if (uniqueApp.size !== appRouteKeys.length) {
  fail('APP_ROUTES contains duplicate route keys')
}

for (const key of uniqueApp) {
  if (!uniqueMeta.has(key)) {
    fail(`APP_ROUTES entry ${key} is missing from ROUTE_META`)
  }
}

for (const key of uniqueMeta) {
  if (!uniqueApp.has(key)) {
    fail(`ROUTE_META entry ${key} is missing from APP_ROUTES`)
  }
}

for (const key of uniqueNav) {
  if (!uniqueMeta.has(key)) {
    fail(`NAV_GROUP_LAYOUT references ${key} which is missing from ROUTE_META`)
  }
}

const lazyImports = [...routesSource.matchAll(/lazy: \(\) => import\('(@\/routes\/[^']+)'\)/g)].map(
  (match) => match[1],
)

for (const importPath of lazyImports) {
  const relativePath = importPath.replace('@/', '') + '.tsx'
  const pagePath = join(srcRoot, relativePath)
  if (!existsSync(pagePath)) {
    fail(`APP_ROUTES lazy import target not found: ${relativePath}`)
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
