export const minTabWidth = 120
export const maxTabWidth = 220
export const moreButtonWidth = 56

// How many tabs fit into the row at minimum width, without a +N menu.
export function maxTabsThatFit(availableWidth: number): number {
    return Math.max(1, Math.floor(availableWidth / minTabWidth))
}

export interface TabsFit {
    visible: number[]
    overflow: number[]
    tabWidth: number
}

// Splits the tabs into those shown in the row and those in the +N menu. The row never drops
// the active tab: when it would be in the menu, it takes the place of the last visible tab.
export function fitTabs(tabCount: number, activeIndex: number, availableWidth: number): TabsFit {
    const all = Array.from({ length: tabCount }, (_, i) => i)

    if (tabCount * minTabWidth <= availableWidth) {
        return { visible: all, overflow: [], tabWidth: tabWidthFor(tabCount, availableWidth) }
    }

    const visibleCount = Math.max(1, Math.floor((availableWidth - moreButtonWidth) / minTabWidth))
    const visible = all.slice(0, visibleCount)
    if (activeIndex >= visibleCount) {
        visible[visibleCount - 1] = activeIndex
    }
    visible.sort((a, b) => a - b)

    return {
        visible,
        overflow: all.filter(i => !visible.includes(i)),
        tabWidth: tabWidthFor(visibleCount, availableWidth - moreButtonWidth),
    }
}

function tabWidthFor(count: number, width: number): number {
    return Math.min(maxTabWidth, Math.max(minTabWidth, width / count))
}
