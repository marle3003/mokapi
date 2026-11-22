import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'
import { compareShots, findImages, findPlaceholders, type CheckReport } from './check.ts'
import { docsDir } from './paths.ts'
import { shots } from './shots.ts'

function markdownFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name)
    if (entry.isDirectory()) return path === join(docsDir, 'resources') ? [] : markdownFiles(path)
    return entry.name.endsWith('.md') ? [path] : []
  })
}

export function checkDocs(): CheckReport {
  const pages = markdownFiles(docsDir).map((path) => ({
    file: relative(docsDir, path),
    markdown: readFileSync(path, 'utf8'),
  }))
  return compareShots(
    shots,
    pages.flatMap((page) => findPlaceholders(page.markdown, page.file)),
    pages.flatMap((page) => findImages(page.markdown, page.file)),
  )
}

/** Prints the findings and returns the exit code. */
export function printReport(report: CheckReport): number {
  const sections: [string, string[]][] = [
    ['Placeholders without screenshot', report.placeholders.map((p) => `${p.file}:${p.line}  ${p.description}`)],
    ['Dashboard images without shot', report.imagesWithoutShot.map((i) => `${i.file}:${i.line}  ${i.src}`)],
    ['Shots without reference in the docs', report.shotsWithoutReference.map((s) => `${s.name}  ${s.file}`)],
  ]
  const findings = sections.filter(([, lines]) => lines.length > 0)
  for (const [title, lines] of findings) {
    console.log(`${title} (${lines.length}):`)
    lines.forEach((line) => console.log(`  ${line}`))
  }
  if (findings.length === 0) console.log('Docs and shots are consistent.')
  return findings.length === 0 ? 0 : 1
}
