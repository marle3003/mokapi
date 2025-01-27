interface DocConfig{
  [name: string]: DocEntry
}

interface DocEntry {
  expanded?: boolean
  hideInNavigation: boolean
  component?: string
  index?: DocEntry
  items?: {[name: string]: string | DocEntry }
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