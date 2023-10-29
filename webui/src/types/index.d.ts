interface DocConfig{
  [name: string]: string | DocConfig | DocEntry
}

interface DocEntry {
  hideInNavigation: boolean
  component?: string
  file?: string
}