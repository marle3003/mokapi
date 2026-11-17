import { spawn, type ChildProcess } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const demoConfigsDir = fileURLToPath(new URL('./demo-configs', import.meta.url))

let proc: ChildProcess | undefined

type StartOptions = {
  /** the Mokapi binary, `mokapi` from the PATH by default */
  bin?: string
  /** no Mokapi output in the terminal */
  quiet?: boolean
  /** extra time after the start, e.g. to get a useful memory usage metric */
  warmUpMs?: number
}

// Starts Mokapi with the demo configs and waits until all demo services are listed. Spawns the
// binary itself (not `go run`), so SIGTERM reaches the process that holds the ports.
export async function startMokapi({ bin = 'mokapi', quiet = false, warmUpMs = 0 }: StartOptions = {}) {
  // MOKAPI_* variables are Mokapi configuration, so MOKAPI_BIN must not reach the process
  const env = { ...process.env }
  delete env.MOKAPI_BIN
  proc = spawn(bin, ['./'], { cwd: demoConfigsDir, stdio: quiet ? 'ignore' : 'inherit', env })

  await waitForMokapi()
  await sleep(warmUpMs)
}

// Polls instead of sleeping a fixed time.
async function waitUntil(what: string, isReady: () => Promise<boolean>, timeoutMs = 60_000) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    try {
      if (await isReady()) return
    } catch {
      // not listening yet
    }
    await sleep(250)
  }
  throw new Error(`${what} not ready after ${timeoutMs / 1000} s`)
}

export const waitForUrl = (url: string) => waitUntil(url, async () => (await fetch(url)).ok)

// The API answers before the configs are loaded, so wait until all demo services are listed.
// Keep the number in sync with the services defined in demo-configs/ (Petstore, Kafka orders,
// MQTT, WebSocket chat, mail, LDAP).
const DEMO_SERVICE_COUNT = 6
const waitForMokapi = () => waitUntil('Mokapi demo services', async () => {
  const services = await (await fetch('http://localhost:8080/api/services')).json()
  return services.length >= DEMO_SERVICE_COUNT
})

// Safe to call when Mokapi was never started or is already stopped.
export function stopMokapi() {
    if (proc && !proc.killed) {
        console.log('🛑 Stopping Mokapi...')
        proc.kill('SIGTERM')
    }
}

function sleep(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}