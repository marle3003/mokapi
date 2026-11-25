export type DocNavKind = 'headline-link' | 'headline-label' | 'group-link' | 'group-button' | 'leaf' | 'hidden'

// Decides how an entry of the docs sidebar is shown. `actives` is the breadcrumb of the open page.
export function useDocNav(actives: () => DocEntry[] | undefined) {
  function hasChildren(entry: DocEntry) {
    return (entry.items?.length ?? 0) > 0
  }

  function kind(entry: DocEntry): DocNavKind {
    if (!hasChildren(entry)) {
      return entry.hideInNavigation ? 'hidden' : 'leaf'
    }
    if (entry.type === 'headline') {
      return entry.path ? 'headline-link' : 'headline-label'
    }
    return entry.path ? 'group-link' : 'group-button'
  }

  function isActive(entry: DocEntry) {
    return actives()?.includes(entry) ?? false
  }

  function isCurrentPage(entry: DocEntry) {
    const breadcrumb = actives()
    return breadcrumb !== undefined && breadcrumb[breadcrumb.length - 1] === entry
  }

  function isExpanded(entry: DocEntry) {
    return entry.type === 'headline' || isActive(entry) || (entry.expanded ?? false)
  }

  function getId(entry: DocEntry) {
    const s = entry.path ?? entry.label
    // Prefix avoids collisions with ids used elsewhere on the page,
    // e.g. a group labelled "app" would otherwise get id="app" like the Vue root element.
    return 'nav' + (s.startsWith('/') ? '' : '-') + s.replaceAll('/', '-').replaceAll(' ', '-')
  }

  return { kind, isActive, isCurrentPage, isExpanded, getId }
}
