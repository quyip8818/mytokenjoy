import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  getRouteLazyImportPaths,
  getMemberRouteLazyImportPaths,
  validateRouteDefinitions,
} from '../src/config/routes.ts'

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

function resolveLazyPagePath(importPath: string): string | null {
  const relativePath = `${importPath.replace('@/', '')}.tsx`
  const indexRelativePath = `${importPath.replace('@/', '')}/index.tsx`
  const pagePath = join(srcRoot, relativePath)
  const indexPagePath = join(srcRoot, indexRelativePath)
  return existsSync(pagePath) ? pagePath : existsSync(indexPagePath) ? indexPagePath : null
}

function assertRegisteredPagesImportFeatureHook(importPaths: string[], label: string) {
  for (const importPath of importPaths) {
    const resolvedPagePath = resolveLazyPagePath(importPath)
    if (!resolvedPagePath) {
      fail(`${label} lazy import target not found: ${importPath}`)
    }

    const pageSource = readFileSync(resolvedPagePath, 'utf8')
    const hasPageHookImport = /from ['"]@\/features\/[^'"]+/.test(pageSource)
    if (!hasPageHookImport) {
      fail(
        `Page ${relativeToSrc(resolvedPagePath)} must import a hook from @/features/ (use-*-page.ts pattern)`,
      )
    }
  }
}

assertRegisteredPagesImportFeatureHook(getRouteLazyImportPaths(), 'ROUTE_DEFINITIONS')
assertRegisteredPagesImportFeatureHook(getMemberRouteLazyImportPaths(), 'MEMBER_ROUTE_DEFINITIONS')

const pageShellExemptPaths = new Set(['routes/auth/login.tsx'])
const routeHookSpreadExemptPaths = new Set(['routes/auth/login.tsx'])
const pageShellWrapperPattern = /\b(PageShell|FilteredPageShell|[A-Z]\w*PageShell)\b/
const routeHookSpreadPattern = /\{\.\.\.use\w+\(\)\}/
const crossFeatureLibImportPattern = /from ['"]@\/features\/([^/'"]+)\/lib\//
const crossFeatureComponentImportPattern = /from ['"]@\/features\/([^/'"]+)\/components\//
const selfFeatureBarrelImportPattern = /from ['"]@\/features\/([^/'"]+)['"]/
const selfFeatureHooksImportPattern = /from ['"]@\/features\/([^/'"]+)\/hooks\//

for (const importPath of [...getRouteLazyImportPaths(), ...getMemberRouteLazyImportPaths()]) {
  const resolvedPagePath = resolveLazyPagePath(importPath)
  if (!resolvedPagePath) continue
  const relativePath = relativeToSrc(resolvedPagePath)
  if (pageShellExemptPaths.has(relativePath)) continue
  const pageSource = readFileSync(resolvedPagePath, 'utf8')
  if (!pageShellWrapperPattern.test(pageSource)) {
    fail(`${relativePath} must use PageShell or an approved layout wrapper`)
  }
  if (!routeHookSpreadExemptPaths.has(relativePath) && !routeHookSpreadPattern.test(pageSource)) {
    fail(`${relativePath} must spread a page hook: {...useX()}`)
  }
}

walkFiles(join(srcRoot, 'features'), (filePath) => {
  const relativePath = relativeToSrc(filePath)
  if (!relativePath.includes('/components/')) return
  const fileName = relativePath.split('/').pop() ?? ''
  if (/page-content\.tsx$/i.test(fileName)) {
    fail(`${relativePath}: use *PageShell naming instead of *PageContent`)
  }
})

const registeredMemberPages = new Set(
  getMemberRouteLazyImportPaths()
    .map((importPath) => resolveLazyPagePath(importPath))
    .filter((pagePath): pagePath is string => pagePath !== null),
)

walkFiles(join(srcRoot, 'routes', 'member'), (filePath) => {
  if (!filePath.endsWith('.tsx')) return
  const relativePath = relativeToSrc(filePath)
  if (relativePath.includes('/hooks/') || relativePath.includes('/components/')) return
  if (!registeredMemberPages.has(filePath)) {
    fail(`${relativePath} is not registered in MEMBER_ROUTE_DEFINITIONS (orphan member page)`)
  }
})

const registeredAdminPages = new Set(
  getRouteLazyImportPaths()
    .map((importPath) => resolveLazyPagePath(importPath))
    .filter((pagePath): pagePath is string => pagePath !== null),
)

const routesDir = join(srcRoot, 'routes')
const orphanSkipPrefixes = ['routes/auth/', 'routes/member/', 'routes/account/']

walkFiles(routesDir, (filePath) => {
  if (!filePath.endsWith('.tsx')) return
  const relativePath = relativeToSrc(filePath)
  if (relativePath.includes('/hooks/') || relativePath.includes('/components/')) return
  if (orphanSkipPrefixes.some((prefix) => relativePath.startsWith(prefix))) return
  if (!registeredAdminPages.has(filePath)) {
    fail(`${relativePath} is not registered in ROUTE_DEFINITIONS (orphan admin page)`)
  }
})

const testsRoot = join(frontendRoot, 'tests')
if (existsSync(join(testsRoot, 'routes'))) {
  fail('tests/routes/ is deprecated; move hook tests to tests/features/{domain}/')
}

walkFiles(routesDir, (filePath) => {
  const relativePath = relativeToSrc(filePath)
  if (relativePath.includes('/hooks/') || relativePath.includes('/components/')) {
    fail(`${relativePath} is deprecated; migrate to features/{domain}/hooks or components/`)
  }
})

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
const legacyDomainComponentDirs = ['budget', 'org', 'keys', 'audit', 'models']

for (const domain of legacyDomainComponentDirs) {
  const domainDir = join(componentsDir, domain)
  if (existsSync(domainDir)) {
    fail(`components/${domain}/ is deprecated; migrate domain UI to features/${domain}/components/`)
  }
}

function forbidInjectedApisInComponents(filePath: string, scopeLabel: string) {
  const source = readFileSync(filePath, 'utf8')
  if (/\buseInjectedApis\s*\(/.test(source)) {
    fail(
      `${relativeToSrc(filePath)}: ${scopeLabel} must not call useInjectedApis(); lift data fetching to use-*-page.ts`,
    )
  }
}

const componentsSkipInjectedApisCheck = new Set(['ui', 'layout', 'auth'])

walkFiles(componentsDir, (filePath) => {
  const relativePath = relativeToSrc(filePath)
  const topSegment = relativePath.split('/')[1]
  if (componentsSkipInjectedApisCheck.has(topSegment ?? '')) return
  forbidInjectedApisInComponents(filePath, 'components/')
})

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

// Component placement rules:
// - features/{domain}/components: shared within a domain
// - routes/{domain}/components: page-private only
// - components/ui: domain-agnostic primitives
walkFiles(join(srcRoot, 'features'), (filePath) => {
  if (!filePath.includes('/components/')) return
  if (filePath.includes('/features/workflow/')) return
  const source = readFileSync(filePath, 'utf8')
  if (/\buseApis\s*\(/.test(source)) {
    fail(
      `${relativeToSrc(filePath)}: features/*/components/ must not call useApis(); lift data fetching to use-*-page.ts`,
    )
  }
  forbidInjectedApisInComponents(filePath, 'features/*/components/')
})

walkFiles(join(srcRoot, 'features'), (filePath) => {
  const relativePath = relativeToSrc(filePath)
  const importerDomain = relativePath.match(/^features\/([^/]+)\//)?.[1]
  if (!importerDomain) return

  const inComponents = /^features\/([^/]+)\/components\//.test(relativePath)
  const inHooks = /^features\/([^/]+)\/hooks\//.test(relativePath)
  const inComponentsOrHooks = inComponents || inHooks

  const source = readFileSync(filePath, 'utf8')
  for (const line of source.split('\n')) {
    const trimmed = line.trimStart()
    if (!trimmed.startsWith('import')) continue

    const libMatch = trimmed.match(crossFeatureLibImportPattern)
    if (libMatch) {
      const importedDomain = libMatch[1]
      if (importedDomain !== importerDomain) {
        fail(
          `${relativePath}: cross-feature deep lib import @/features/${importedDomain}/lib/ is forbidden; use @/features/${importedDomain} barrel`,
        )
      }
      if (inComponentsOrHooks && importedDomain === importerDomain) {
        fail(
          `${relativePath}: deep lib import @/features/${importedDomain}/lib/ is forbidden in components/hooks; use @/features/${importedDomain} barrel`,
        )
      }
    }

    const componentMatch = trimmed.match(crossFeatureComponentImportPattern)
    if (componentMatch) {
      const importedDomain = componentMatch[1]
      if (importedDomain !== importerDomain) {
        fail(
          `${relativePath}: cross-feature component import @/features/${importedDomain}/components/ is forbidden; use @/features/${importedDomain} barrel`,
        )
      }
    }

    if (inHooks) {
      const barrelMatch = trimmed.match(selfFeatureBarrelImportPattern)
      if (barrelMatch?.[1] === importerDomain) {
        fail(
          `${relativePath}: hooks must not import from @/features/${importerDomain} barrel; use ../lib/ or sibling hook relative paths`,
        )
      }
    }

    if (inComponents) {
      const hooksMatch = trimmed.match(selfFeatureHooksImportPattern)
      if (hooksMatch?.[1] === importerDomain) {
        fail(
          `${relativePath}: components must not import from @/features/${importerDomain}/hooks/; use @/features/${importerDomain} barrel`,
        )
      }
    }
  }
})

const deepRelativeImportPattern = /from ['"]\.\.\/\.\.\/|import\(['"]\.\.\/\.\.\//

walkFiles(srcRoot, (filePath) => {
  const source = readFileSync(filePath, 'utf8')
  for (const line of source.split('\n')) {
    const trimmed = line.trimStart()
    if (!deepRelativeImportPattern.test(trimmed)) continue
    fail(
      `${relativeToSrc(filePath)}: use @/ alias instead of ../../ (or deeper) relative imports in src/`,
    )
  }
})

console.log('check-conventions: all checks passed')
