import { parseArgs } from 'node:util'
import { checkDocs, printReport } from './docs-check.ts'
import { shots } from './shots.ts'

const { values } = parseArgs({
  options: {
    only: { type: 'string' },
    theme: { type: 'string', default: 'dark' },
    check: { type: 'boolean', default: false },
  },
})

if (values.theme !== 'dark' && values.theme !== 'light') {
  console.error(`Unknown --theme ${values.theme}, use dark or light`)
  process.exitCode = 1
} else if (values.check) {
  process.exitCode = printReport(checkDocs())
} else {
  const selected = shots.filter((s) => !values.only || s.name.includes(values.only) || s.file.includes(values.only))
  if (selected.length === 0) {
    console.error(`No shot matches --only ${values.only}`)
    process.exitCode = 1
  } else {
    // the runner loads the demo drivers (Kafka, MQTT, ...), which --check doesn't need
    const { runShots } = await import('./runner.ts')
    process.exitCode = await runShots(selected, values.theme)
  }
}
