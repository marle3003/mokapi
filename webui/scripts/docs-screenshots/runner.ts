import { spawn, execFileSync } from 'node:child_process'
import { join } from 'node:path'
import { chromium, type Browser, type Page } from '@playwright/test'
import { driveAll } from '../dashboard-demo/drive-all.ts'
import { closeKafka } from '../dashboard-demo/drive-kafka.ts'
import { startMokapi, stopMokapi, waitForUrl } from '../dashboard-demo/mokapi.ts'
import { demoDir, docsDir, webuiDir } from './paths.ts'
import type { Shot } from './shots.ts'

const vitePort = 5175
const viteOrigin = `http://localhost:${vitePort}`

// The binary from MOKAPI_BIN (e2e.sh builds it once), otherwise built once into $TMPDIR.
function mokapiBinary(): string {
  if (process.env.MOKAPI_BIN) return process.env.MOKAPI_BIN
  const bin = join(process.env.TMPDIR ?? '/tmp', 'mokapi-docs-screenshots')
  console.log('Building Mokapi...')
  execFileSync('go', ['build', '-o', bin, './cmd/mokapi'], { cwd: join(webuiDir, '..'), stdio: 'inherit' })
  return bin
}

// The dashboard in dashboard mode, its API requests go to the demo Mokapi.
async function startVite(onSpawn: (stop: () => void) => void) {
  // another server on the port would answer the readiness check while our Vite dies on --strictPort
  if (await answers(viteOrigin)) throw new Error(`Port ${vitePort} is already in use, stop the server using it`)
  const vite = spawn('./node_modules/.bin/vite', ['--mode', 'dashboard', '--port', `${vitePort}`, '--strictPort'], {
    cwd: webuiDir,
    stdio: 'ignore',
    env: { ...process.env, MOKAPI_DASHBOARD_API: 'http://localhost:8080' },
  })
  onSpawn(() => vite.kill('SIGTERM')) // registered before the wait, so a failing start doesn't leak Vite
  const exited = new Promise<never>((_, reject) =>
    vite.once('exit', (code) => reject(new Error(`Vite exited with code ${code} before it answered on port ${vitePort}`))),
  )
  await Promise.race([waitForUrl(viteOrigin), exited])
}

async function answers(url: string) {
  try {
    await fetch(url)
    return true
  } catch {
    return false
  }
}

// Relative times and the job countdown stay the same while the screenshot is taken.
async function freezeClock(page: Page) {
  const now = await page.evaluate(() => Date.now())
  await page.clock.pauseAt(now + 100)
}

async function takeShot(browser: Browser, shot: Shot, theme: string, clockStart: number) {
  const context = await browser.newContext({
    baseURL: viteOrigin,
    viewport: shot.viewport ?? { width: 1400, height: 900 },
    deviceScaleFactor: 2,
    colorScheme: theme === 'light' ? 'light' : 'dark',
    storageState: { cookies: [], origins: [{ origin: viteOrigin, localStorage: [{ name: 'theme', value: theme }] }] },
  })
  try {
    const page = await context.newPage()
    await page.clock.install({ time: clockStart })
    await page.goto(shot.url)
    await shot.steps?.(page)
    await freezeClock(page)
    await page.screenshot({ path: join(docsDir, shot.file), clip: await shot.clip?.(page) })
  } finally {
    await context.close()
  }
}

/**
 * Starts Mokapi with the demo data and the dashboard, takes the shots and stops everything again,
 * also on SIGINT and SIGTERM. A failing shot is reported with its name and the others continue.
 * Returns the exit code.
 */
export async function runShots(selected: Shot[], theme: string): Promise<number> {
  const stops: (() => unknown)[] = []
  const stopAll = async () => {
    for (const stop of stops.reverse()) {
      try {
        await stop()
      } catch {
        // the other stops must still run
      }
    }
  }
  const onSignal = () => stopAll().finally(() => process.exit(1))
  process.once('SIGINT', onSignal)
  process.once('SIGTERM', onSignal)

  let failed = 0
  try {
    // registered before the start, so a failing start doesn't leak Mokapi
    stops.push(stopMokapi, closeKafka)
    await startMokapi({ bin: mokapiBinary(), quiet: true })
    process.chdir(demoDir)
    await driveAll()
    process.chdir(webuiDir)
    const clockStart = Date.now() + 2000

    await startVite((stop) => stops.push(stop))
    const browser = await chromium.launch()
    stops.push(() => browser.close())

    for (const shot of selected) {
      try {
        await takeShot(browser, shot, theme, clockStart)
        console.log(`ok      ${shot.name}  ${shot.file}`)
      } catch (err) {
        failed++
        console.error(`FAILED  ${shot.name}  ${shot.file}\n${err instanceof Error ? err.message : err}`)
      }
    }
  } catch (err) {
    console.error(err)
    failed++
  } finally {
    process.off('SIGINT', onSignal)
    process.off('SIGTERM', onSignal)
    await stopAll()
  }
  return failed === 0 ? 0 : 1
}
