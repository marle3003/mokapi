interface DocConfig{
  [name: string]: DocEntry
}

interface DocEntry {
  expanded?: boolean
  hideInNavigation: boolean
  component?: string
  items?: {[name: string]: string | DocEntry }
}