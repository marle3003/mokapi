// @vitest-environment node
import { createLogger, createServer } from 'vite'
import { expect, test } from 'vitest'

// Vite's dependency scan reads imports line by line. An example script inside a template string
// (e.g. in Mail.vue) that starts a line with "import ... from 'mokapi/mail'" is taken as a real
// import. When it can't be resolved the scan fails and Vite skips pre-bundling, which makes the
// dev server re-optimize while pages load ("Failed to fetch dynamically imported module").
test('dependency scan of the dashboard resolves all imports', async () => {
  const warnings: string[] = []
  const logger = createLogger('warn', { allowClearScreen: false })
  logger.warn = (msg) => { warnings.push(msg) }
  logger.error = (msg) => { warnings.push(msg) }

  const server = await createServer({
    mode: 'dashboard',
    configFile: 'vite.config.ts',
    customLogger: logger,
    server: { middlewareMode: true, watch: null, ws: false },
    // the scan must not be influenced by (or write to) a cache of a running dev server
    cacheDir: 'node_modules/.vite-scan-test',
    optimizeDeps: { force: true }
  })
  try {
    await server.environments.client.depsOptimizer?.scanProcessing
  } finally {
    await server.close()
  }

  expect(warnings.filter((w) => w.includes('dependency scan'))).toEqual([])
}, 60000)
