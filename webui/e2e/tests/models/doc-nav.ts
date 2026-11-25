import type { Page } from '@playwright/test'

// The sidebar of the docs: sections (headlines), groups with a chevron and plain entries.
export function useDocNav(page: Page) {
    const sidebar = page.getByRole('navigation', { name: 'Sidebar' })
    return {
        sidebar,
        link: (name: string) => sidebar.getByRole('link', { name, exact: true }),
        // the chevron next to a group name; it only expands and collapses the group
        chevron: (group: string) => sidebar.getByRole('button', { name: `Toggle ${group}`, exact: true }),
        group: (name: string) => sidebar.getByRole('region', { name, exact: true }),
    }
}

// a link of the sidebar is marked active (not the `child-active` class of a group button)
export const activeClass = /(^|\s)active(\s|$)/
