interface DocConfig{
  [name: string]: DocEntry
}

interface DocEntry {
  expanded?: boolean
  hideNavigation: boolean
  hideInNavigation: boolean
  component?: string
  index?: DocEntry
  items?: {[name: string]: string | DocEntry }
}

interface DocMeta {
  title: string
  description: string
  icon: string
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