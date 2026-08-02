import { reactive, readonly, onUnmounted, getCurrentInstance } from 'vue'

export interface ToolbarItem {
    id?: number
    label: string
    icon?: string
    onClick: () => void
}

const items = reactive<ToolbarItem[]>([])
let uid = 0

export function useToolbarItems() {
    return {
        items: readonly(items)
    }
}

export function useToolbarItem(...itemsToAdd: ToolbarItem[]) {
    const ids: number[] = []
    for (const item of itemsToAdd) {
        const id = ++uid
        items.push({ id, ...item })
        ids.push(id)
    }

    // only auto-cleanup if called from within a component's setup()
    if (getCurrentInstance()) {
        onUnmounted(() => {
            for (const id of ids) {
                const idx = items.findIndex(i => i.id === id)
                if (idx !== -1) items.splice(idx, 1)
            }
        })
    }

    return ids
}

/**
 * Manual removal, for cases where an item's visibility depends on
 * conditions other than "component is mounted" (e.g. a v-if in the
 * caller). useToolbarItem's auto-cleanup covers the common case.
 */
export function removeToolbarItem(id: number) {
    const idx = items.findIndex(i => i.id === id)
    if (idx !== -1) items.splice(idx, 1)
}