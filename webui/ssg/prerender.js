// Pre-renders the website after `vite build --mode website` and the server build
// (`vite build --mode website --ssr src/entry-server.ts --outDir dist-ssr`):
// `node ssg/prerender.js [distDir]` (distDir defaults to ../dist, the server bundle is expected in
// the folder `dist-ssr` next to it).
//
// The built index.html is the template. The app is rendered with its content and head into
// <div id="app"> and written to `<path>.html` (`/` as home.html) for
//   - `/` and every URL of sitemap.xml,
//   - every internal link target in a generated page that is a route of the app (e.g. /resources,
//     which is not in the sitemap), until no new target appears. Not /dashboard-demo (rendered in the
//     browser) and not redirects of .htaccess.
// The web server serves it for `/path` (public/website/.htaccess), the browser hydrates it.
// Fails (exit code 1) when
//   - a route is "page not found",
//   - a URL of sitemap.xml has no generated file,
//   - a generated page has no <h1> (the page not found view has none either),
//   - public/website/.htaccess still switches on HTTP_USER_AGENT (every visitor gets the same files).
// Internal links without a target are only printed.
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const webuiDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '..')
const htaccessFile = path.join(webuiDir, 'public/website/.htaccess')
const siteOrigin = 'https://mokapi.io'

// rendered in the browser, from the demo recording
const demoDashboardRoute = '/dashboard-demo'

export function sitemapPaths(distDir) {
  const sitemap = fs.readFileSync(path.join(distDir, 'sitemap.xml'), 'utf8')
  return [...sitemap.matchAll(/<loc>(.*?)<\/loc>/g)].map(([, url]) => new URL(url).pathname)
}

// `/` is the shell of the app (index.html), the home page is generated as home.html
export function generatedFileOf(distDir, urlPath) {
  const name = urlPath === '/' ? 'home' : decodeURIComponent(urlPath).replace(/^\/|\/$/g, '')
  return path.join(distDir, name + '.html')
}

export function checkSitemapHasFiles(distDir, paths) {
  return paths
    .filter((p) => !fs.existsSync(generatedFileOf(distDir, p)))
    .map((p) => `no file for ${p}: expected ${path.relative(distDir, generatedFileOf(distDir, p))}`)
}

export function checkPageHasContent(distDir, paths) {
  const errors = []
  for (const p of paths) {
    const file = generatedFileOf(distDir, p)
    if (!fs.existsSync(file)) {
      continue // reported by checkSitemapHasFiles
    }
    if (!/<h1[\s>]/.test(fs.readFileSync(file, 'utf8'))) {
      errors.push(`${p} has no <h1> (the page not found view has none either)`)
    }
  }
  return errors
}

export function checkHtaccessIsNotUserAgentBased(file = htaccessFile) {
  return fs.readFileSync(file, 'utf8').includes('HTTP_USER_AGENT')
    ? [`${path.relative(webuiDir, file)} still depends on HTTP_USER_AGENT`]
    : []
}

// Source patterns of the unconditional redirects in .htaccess, e.g. ^docs/guides/http/?$
// (rules after a RewriteCond, like www to non-www, don't say which pages are redirected)
function redirectPatterns(file = htaccessFile) {
  const patterns = []
  let conditioned = false
  for (const line of fs.readFileSync(file, 'utf8').split('\n')) {
    if (/^\s*RewriteCond\s/.test(line)) {
      conditioned = true
    } else if (/^\s*RewriteRule\s/.test(line)) {
      const pattern = line.match(/^\s*RewriteRule\s+(\S+)\s+\S+\s+\[[^\]]*R=30\d/)?.[1]
      if (pattern && !conditioned) {
        patterns.push(new RegExp(pattern))
      }
      conditioned = false
    }
  }
  return patterns
}

function internalLinksOf(html) {
  return [...html.matchAll(/<a\s[^>]*href="([^"]*)"/g)]
    .map(([, href]) => {
      try {
        return new URL(href.replace(/&amp;/g, '&'), siteOrigin + '/')
      } catch {
        return undefined
      }
    })
    .filter((url) => url !== undefined && url.origin === siteOrigin)
    .map((url) => decodeURIComponent(url.pathname))
}

function isFile(file) {
  return fs.existsSync(file) && fs.statSync(file).isFile()
}

function isDemoDashboard(link) {
  return link === demoDashboardRoute || link.startsWith(demoDashboardRoute + '/')
}

function isRedirect(link, redirects) {
  return redirects.some((pattern) => pattern.test(link.replace(/^\//, '')))
}

export function findLinksWithoutTarget(distDir, paths) {
  const redirects = redirectPatterns()
  const hasTarget = (link) =>
    link === '/' ||
    fs.existsSync(generatedFileOf(distDir, link)) ||
    isFile(path.join(distDir, link)) ||
    isDemoDashboard(link) ||
    isRedirect(link, redirects)

  const warnings = []
  for (const p of new Set(['/', ...paths])) {
    const file = generatedFileOf(distDir, p)
    if (!fs.existsSync(file)) {
      continue
    }
    for (const link of new Set(internalLinksOf(fs.readFileSync(file, 'utf8')))) {
      if (!hasTarget(link)) {
        warnings.push(`${p} links to ${link}`)
      }
    }
  }
  return warnings
}

function escapeHtml(text) {
  return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

// A head tag of the page replaces the tag of the template with the same selector, e.g.
// meta[name="description"] or script[type="application/ld+json"] (see HeadTag in meta.ts).
function withoutTag(html, { selector }) {
  const [, tagName, attribute, value] = selector.match(/^(\w+)\[([\w:-]+)="([^"]*)"\]$/)
  const attributeValue = value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const opening = `<${tagName}\\s[^>]*\\b${attribute}="${attributeValue}"[^>]*>`
  const element = tagName === 'script' ? `${opening}[\\s\\S]*?</script>` : opening
  return html.replace(new RegExp(`[ \\t]*${element}\\n?`, 'i'), '')
}

function renderTag({ tagName, attributes, text }) {
  const attributeList = Object.entries(attributes).map(([name, value]) => ` ${name}="${escapeHtml(value)}"`).join('')
  // JSON must not be escaped, but must not end the script element either
  return tagName === 'script'
    ? `<script${attributeList}>${text.replace(/</g, '\\u003c')}</script>`
    : `<${tagName}${attributeList}>`
}

export function renderPage(template, { html, head }) {
  let page = template
  if (head.title !== undefined) {
    page = page.replace(/<title>[^<]*<\/title>/, () => `<title>${escapeHtml(head.title)}</title>`)
  }
  for (const tag of head.tags) {
    page = withoutTag(page, tag)
  }
  const tags = head.tags.map((tag) => `    ${renderTag(tag)}\n`).join('')
  return page
    .replace('</head>', () => `${tags}  </head>`)
    .replace('<div id="app"></div>', () => `<div id="app">${html}</div>`)
}

export async function loadServerBundle(distDir) {
  return import(pathToFileURL(path.join(distDir, '..', 'dist-ssr', 'entry-server.js')).href)
}

// Returns the rendered paths, the ones found through links (not in sitemap.xml) and the errors.
export async function prerender(distDir, { render, isRouteOfApp }) {
  const template = fs.readFileSync(path.join(distDir, 'index.html'), 'utf8')
  const redirects = redirectPatterns()
  const errors = []
  const rendered = []
  const linked = []
  const queued = new Set(['/', ...sitemapPaths(distDir)])
  const isNewPageOfApp = (link) =>
    !queued.has(link) && !isDemoDashboard(link) && !isRedirect(link, redirects) && isRouteOfApp(link)

  // one after the other: router and Markdown cache of the app are singletons
  for (const urlPath of queued) {
    const page = await render(urlPath)
    if (page.routeName === 'not-found') {
      errors.push(`${urlPath} is not a route of the app (not-found)`)
      continue
    }
    const file = generatedFileOf(distDir, urlPath)
    const content = renderPage(template, page)
    fs.mkdirSync(path.dirname(file), { recursive: true })
    fs.writeFileSync(file, content)
    rendered.push(urlPath)

    for (const link of internalLinksOf(content)) {
      if (isNewPageOfApp(link)) {
        queued.add(link) // a Set visits what is added while looping
        linked.push(link)
      }
    }
  }
  return { rendered, linked, errors }
}

// `paths` are the pages that were rendered
export function runChecks(distDir, paths = ['/', ...sitemapPaths(distDir)]) {
  const errors = [
    ...checkSitemapHasFiles(distDir, sitemapPaths(distDir)),
    ...checkPageHasContent(distDir, paths),
    ...checkHtaccessIsNotUserAgentBased(),
  ]
  const warnings = findLinksWithoutTarget(distDir, paths)
  return { errors, warnings }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const distDir = path.resolve(process.argv[2] ?? path.join(webuiDir, 'dist'))
  const serverBundle = await loadServerBundle(distDir)
  const { rendered, linked, errors: renderErrors } = await prerender(distDir, serverBundle)
  const { errors, warnings } = runChecks(distDir, rendered)
  errors.unshift(...renderErrors)
  console.log(`prerender: ${rendered.length} pages rendered, ${linked.length} of them found through links`)
  for (const link of linked) {
    console.log(`  rendered from links: ${link}`)
  }
  for (const warning of warnings) {
    console.warn(`warning: ${warning}`)
  }
  for (const error of errors) {
    console.error(`error: ${error}`)
  }
  if (errors.length > 0) {
    console.error(`prerender: ${errors.length} check(s) failed`)
    process.exit(1)
  }
  console.log(`prerender: all checks passed (${warnings.length} warning(s))`)
}
