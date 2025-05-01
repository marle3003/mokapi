interface DocConfig{
  [name: string]: DocEntry
}

interface DocEntry {
  expanded?: boolean
  hideNavigation: boolean
  hideInNavigation: boolean
  canonical?: string
  component?: string
  index?: DocEntry
  items?: {[name: string]: string | DocEntry }
}

interface DocMeta {
  title: string
  description: string
  icon: string | undefined
  tech: string | undefined
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
  duration: number
  tags: { [name: string]: string}
  logs: { level: string, message: string}[]
  error?: { message: string }
}