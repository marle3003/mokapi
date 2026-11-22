interface DocConfig {
  [name: string]: DocEntry
}

interface DocEntry {
  label: string
  source?: string
  path?: string
  index?: DocEntry
  items?: DocEntry[]
  type?: 'headline' | 'root'

  expanded?: boolean
  hideNavigation?: boolean
  hideInNavigation?: boolean
  canonical?: string
  
  component?: string

  title?: string
  description?: string
}

interface DocMeta {
  title: string
  description: string
  icon?: string
  tech?: string
  image?: { url: string, alt: string }
  subtitle?: string
  tags?: string[]
}

interface Source {
  preview?: Data
  binary?: Data
}

interface Data {
  content: string
  contentType: string
  contentTypeTitle?: string
  description?: string
}

interface JobExecution {
  schedule: string
  maxRuns: number
  runs: number
  nextRun: string
  duration: number
  tags: { [name: string]: string}
  logs: { level: string, message: string}[]
  error?: { message: string }
}

// A registered job of `GET /api/jobs`. `maxRuns` is -1 for a job without a limit, `nextRun` and
// `lastRun` are missing if there is none.
declare interface JobInfo {
  id: string
  name: string
  schedule: string
  maxRuns: number
  runs: number
  state: 'scheduled' | 'finished' | 'cancelled'
  nextRun?: string
  tags: { [name: string]: string }
  lastRun?: JobLastRun
}

declare interface JobLastRun {
  eventId: string
  time: string
  duration: number
  status: 'success' | 'warning' | 'error'
}

declare interface BreadcrumbItem {
  text: string
  to?: RouteLocationRaw
}