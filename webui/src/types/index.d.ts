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