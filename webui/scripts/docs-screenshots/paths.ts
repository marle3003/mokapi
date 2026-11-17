import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

export const webuiDir = fileURLToPath(new URL('../..', import.meta.url))
export const docsDir = join(webuiDir, '..', 'docs')
// the drivers read ./assets/... relative to this folder, like ci.ts
export const demoDir = join(webuiDir, 'scripts', 'dashboard-demo')
