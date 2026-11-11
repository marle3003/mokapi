// Tests public/website/.htaccess against a real Apache: `node ssg/test-htaccess.js [distDir]`
// (npm run test-htaccess, after npm run build-website; distDir defaults to ../dist).
//
// Starts httpd:2.4 in Docker (ssg/httpd-test.conf) with the built website as document root, requests
// every URL of the table below and checks status, redirect target and the served file. Fails (exit
// code 1) when a check fails, when an unconditional redirect of .htaccess has no URL in the table,
// or when Docker is not reachable. The container is removed at the end and on SIGINT/SIGTERM.
import fs from 'node:fs'
import http from 'node:http'
import path from 'node:path'
import { execFileSync, spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { redirectPatterns } from './prerender.js'

const ssgDir = path.dirname(fileURLToPath(import.meta.url))
const image = 'httpd:2.4'
const container = `mokapi-htaccess-${process.pid}`

// The "https" vhost (SERVER_PORT 443) is where all rules but the http to https redirect are tested.
// Entries: { path, status, location?, file?, host?, vhost?, accept? }
//   location: a path (compared without scheme, host and port, which Apache builds from the plain
//             HTTP vhost) or a full URL
//   file:     the body must be equal to this file of dist
const notFound = { status: 404, file: '404.html' }
export const urls = [
  // pages
  { path: '/', status: 200, file: 'home.html' },
  { path: '/docs/http/overview', status: 200, file: 'docs/http/overview.html' },
  { path: '/docs', status: 200, file: 'docs.html' },
  { path: '/resources', status: 200, file: 'resources.html' },
  { path: '/http', status: 200, file: 'http.html' },
  { path: '/home', status: 200, file: 'home.html' },
  { path: '/icon.svg', status: 200, file: 'icon.svg' },
  { path: '/404.html', status: 200, file: '404.html' },

  // unknown URLs and missing files: 404 with the not found page
  { path: '/does-not-exist', ...notFound },
  { path: '/docs/foo', ...notFound },
  { path: '/resources/foo', ...notFound },
  { path: '/resources/blogs/foo', ...notFound },
  { path: '/docs/http', ...notFound }, // a folder without a page of its own
  { path: '/assets', ...notFound },
  { path: '/missing.png', ...notFound },
  { path: '/assets/missing.js', ...notFound },
  { path: '/docs/missing.html', ...notFound },

  // the demo dashboard is rendered in the browser
  { path: '/dashboard-demo', status: 200, file: 'index.html' },
  { path: '/dashboard-demo/http/services/x', status: 200, file: 'index.html' },
  { path: '/dashboard-demo/missing.js', accept: '*/*', ...notFound },

  // trailing slash
  { path: '/docs/http/overview/', status: 301, location: '/docs/http/overview' },
  { path: '/resources/', status: 301, location: '/resources' },
  { path: '/http/', status: 301, location: '/http' },
  { path: '/docs/', status: 301, location: '/docs' },
  { path: '/docs/?a=1', status: 301, location: '/docs?a=1' },
  { path: '/docs/http/', status: 301, location: '/docs/http' },
  { path: '/does-not-exist/', status: 301, location: '/does-not-exist' },
  { path: '/dashboard-demo/', status: 301, location: '/dashboard-demo' },

  // redirects
  { path: '/docs/examples', status: 301, location: '/resources' },
  { path: '/docs/examples/', status: 301, location: '/resources' },
  { path: '/docs/examples/foo', status: 301, location: '/resources/foo' },
  { path: '/docs/blogs', status: 301, location: '/resources/blogs' },
  { path: '/docs/blogs/foo', status: 301, location: '/resources/blogs/foo' },
  { path: '/docs/resources/foo', status: 301, location: '/resources/foo' },
  { path: '/docs/guides/http', status: 301, location: '/docs/http/overview' },
  { path: '/docs/guides/kafka', status: 301, location: '/docs/kafka/overview' },
  { path: '/docs/guides/ldap', status: 301, location: '/docs/ldap/overview' },
  { path: '/docs/guides/mail', status: 301, location: '/docs/mail/overview' },
  { path: '/docs/guides/foo', status: 301, location: '/docs/foo' },
  { path: '/smtp', status: 301, location: '/mail' },
  { path: '/docs/smtp/foo', status: 301, location: '/docs/mail/foo' },
  { path: '/docs/http/https-&-tls', status: 301, location: '/docs/http/tls' },
  { path: '/resources/blogs/ensuring-api-contract-compliance-with-mokapi', status: 301, location: '/resources/blogs/guard-your-api-contracts' },
  { path: '/resources/blogs/acceptance-testing-with-mokapi:-focus-on-what-matters', status: 301, location: '/resources/blogs/acceptance-testing' },

  // www and http
  { path: '/docs/http/overview', host: 'www.mokapi.io', status: 301, location: 'https://mokapi.io/docs/http/overview' },
  { path: '/docs/http/overview', vhost: 'http', status: 301, location: 'https://mokapi.io/docs/http/overview' },
  { path: '/docs/http/overview', vhost: 'http', host: 'www.mokapi.io', status: 301, location: 'https://mokapi.io/docs/http/overview' },
]

// Every unconditional redirect of .htaccess needs a URL in the table that it redirects, so a new
// redirect without a test fails. (The conditional ones, www and http, are in the table above.)
export function redirectsWithoutUrl(table = urls, patterns = redirectPatterns()) {
  return patterns
    .filter((pattern) => !table.some((entry) => entry.status === 301 && pattern.test(entry.path.replace(/^\//, ''))))
    .map((pattern) => `no URL in the table for the redirect ${pattern}`)
}

// Apache builds the Location of a redirect from the plain HTTP virtual host, e.g.
// http://mokapi.io:443/resources: a path is compared without scheme, host and port.
export function locationMatches(expected, actual) {
  if (actual === undefined) {
    return false
  }
  if (!expected.startsWith('/')) {
    return actual === expected
  }
  const url = new URL(actual, 'http://mokapi.io')
  return url.pathname + url.search === expected
}

function request(port, { path: urlPath, host = 'mokapi.io', accept = 'text/html' }) {
  return new Promise((resolve, reject) => {
    // node:http, because fetch doesn't let a request choose its Host header
    const req = http.request({ host: '127.0.0.1', port, path: urlPath, headers: { Host: host, Accept: accept } }, (res) => {
      const chunks = []
      res.on('data', (chunk) => chunks.push(chunk))
      res.on('end', () => resolve({ status: res.statusCode, location: res.headers.location, body: Buffer.concat(chunks) }))
    })
    req.on('error', reject)
    // a port that Docker already publishes may accept the connection before Apache answers
    req.setTimeout(2000, () => req.destroy(new Error('no answer within 2 s')))
    req.end()
  })
}

export async function check(port, distDir, entry) {
  const response = await request(port, entry)
  const errors = []
  if (response.status !== entry.status) {
    errors.push(`status ${response.status}, expected ${entry.status}`)
  }
  if (entry.location !== undefined && !locationMatches(entry.location, response.location)) {
    errors.push(`Location ${response.location}, expected ${entry.location}`)
  }
  if (entry.file !== undefined) {
    const file = path.join(distDir, entry.file)
    if (!fs.existsSync(file)) {
      errors.push(`dist has no file ${entry.file}`)
    } else if (!response.body.equals(fs.readFileSync(file))) {
      errors.push(`body is not the file ${entry.file}`)
    }
  }
  return errors
}

function docker(...args) {
  return execFileSync('docker', args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim()
}

function failWithoutDocker() {
  const ps = spawnSync('docker', ['ps', '-q'], { encoding: 'utf8' })
  if (ps.status !== 0) {
    const reason = ps.error?.message ?? ps.stderr.trim().split('\n')[0]
    console.error(`test-htaccess: Docker is not available (${reason}). Start Docker and run again: the test runs ${image}.`)
    process.exit(1)
  }
}

function hostPort(containerPort) {
  const port = Number(docker('port', container, `${containerPort}/tcp`).split('\n')[0].split(':').pop())
  if (!port) {
    // right after docker run, Docker Desktop may not have published the port yet
    throw new Error(`port ${containerPort} is not published yet`)
  }
  return port
}

async function waitUntilServing() {
  let lastError
  for (let i = 0; i < 60; i++) {
    try {
      const ports = { http: hostPort(80), https: hostPort(443) }
      await request(ports.https, { path: '/' })
      return ports
    } catch (error) {
      lastError = error
      await new Promise((resolve) => setTimeout(resolve, 500))
    }
  }
  throw new Error(`Apache does not answer (${lastError?.message}), log of the container:\n${containerLog()}`)
}

function containerLog() {
  // httpd logs to stderr
  const logs = spawnSync('docker', ['logs', container], { encoding: 'utf8' })
  return `${logs.stdout}${logs.stderr}`.trim()
}

function removeContainer() {
  try {
    docker('rm', '-f', container)
  } catch {
    // not started
  }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const distDir = path.resolve(process.argv[2] ?? path.join(ssgDir, '..', 'dist'))
  if (!fs.existsSync(path.join(distDir, '.htaccess'))) {
    console.error(`test-htaccess: ${distDir} has no .htaccess, run npm run build-website first`)
    process.exit(1)
  }
  failWithoutDocker()

  for (const signal of ['SIGINT', 'SIGTERM']) {
    process.on(signal, () => {
      removeContainer()
      process.exit(1)
    })
  }

  const failures = redirectsWithoutUrl()
  try {
    docker('run', '--rm', '-d', '--name', container, '-p', '127.0.0.1::80', '-p', '127.0.0.1::443',
      '-v', `${distDir}:/usr/local/apache2/htdocs:ro`,
      '-v', `${path.join(ssgDir, 'httpd-test.conf')}:/usr/local/apache2/conf/httpd.conf:ro`, image)
    const ports = await waitUntilServing()

    for (const entry of urls) {
      const vhost = entry.vhost ?? 'https'
      // a failing request is a failure of this URL, the others are still checked
      const errors = await check(ports[vhost], distDir, entry).catch((error) => [error.message || String(error)])
      for (const error of errors) {
        failures.push(`${vhost} ${entry.host ?? ''} ${entry.path}: ${error}`.replace(/\s+/g, ' '))
      }
    }
  } finally {
    removeContainer()
  }

  for (const failure of failures) {
    console.error(`error: ${failure}`)
  }
  if (failures.length > 0) {
    console.error(`test-htaccess: ${failures.length} check(s) failed`)
    process.exit(1)
  }
  console.log(`test-htaccess: ${urls.length} URLs checked, all redirects of .htaccess covered`)
}
