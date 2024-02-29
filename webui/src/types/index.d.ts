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