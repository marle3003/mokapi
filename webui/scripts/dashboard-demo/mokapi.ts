import { spawn, type ChildProcess } from 'child_process'

let proc: ChildProcess

export async function startMokapi() {
  proc = spawn('mokapi', ['./demo-configs'], {
    stdio: 'inherit',
    env: {
      ...process.env,
    }
  })

  // sleep at least 5sec to get useful memory usage metric
  await sleep(5500);
  await fetch('http://localhost:8080')
}

export function stopMokapi() {
    if (!proc.killed) {
        console.log('🛑 Stopping Mokapi...')
        proc.kill('SIGTERM')
    }
}

function sleep(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}