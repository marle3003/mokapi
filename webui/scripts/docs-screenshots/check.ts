export type Placeholder = { file: string; line: number; description: string }
export type ImageRef = { file: string; line: number; src: string }
export type ShotTarget = { name: string; file: string }
export type CheckReport = {
  placeholders: Placeholder[]
  imagesWithoutShot: ImageRef[]
  shotsWithoutReference: ShotTarget[]
}

// Pages whose images show the dashboard (a file name prefix is not enough: petstore-rest-api-*.jpg).
export const DASHBOARD_PAGES = [
  'get-started/dashboard.md',
  'get-started/running.md',
  ...['http', 'kafka', 'mqtt', 'websocket', 'ldap', 'mail'].map((area) => `${area}/dashboard.md`),
]

const PLACEHOLDER = /<!--\s*screenshot:\s*(.+?)\s*-->/g
const IMAGE = /<img\b[^>]*?\bsrc="\/docs\/([^"]+)"/g

function scanLines<T>(markdown: string, pattern: RegExp, toResult: (match: RegExpExecArray, line: number) => T): T[] {
  return markdown.split('\n').flatMap((text, index) =>
    [...text.matchAll(pattern)].map((match) => toResult(match, index + 1)),
  )
}

export function findPlaceholders(markdown: string, file: string): Placeholder[] {
  return scanLines(markdown, PLACEHOLDER, (match, line) => ({ file, line, description: match[1] }))
}

export function findImages(markdown: string, file: string): ImageRef[] {
  return scanLines(markdown, IMAGE, (match, line) => ({ file, line, src: match[1] }))
}

export function compareShots(shots: ShotTarget[], placeholders: Placeholder[], images: ImageRef[]): CheckReport {
  const isShotTarget = (src: string) => shots.some((shot) => shot.file === src)
  return {
    placeholders,
    imagesWithoutShot: images.filter((image) => DASHBOARD_PAGES.includes(image.file) && !isShotTarget(image.src)),
    shotsWithoutReference: shots.filter((shot) => !images.some((image) => image.src === shot.file)),
  }
}
