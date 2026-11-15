import { onUnmounted, ref } from 'vue'

// Bootstrap breakpoint md: below 768 px the sidebar is an overlay.
// Only for the dashboard, which is not rendered on the server.
export function useNarrowWindow() {
    const query = window.matchMedia('(max-width: 767.98px)')
    const isNarrow = ref(query.matches)
    const update = (e: MediaQueryListEvent) => { isNarrow.value = e.matches }

    query.addEventListener('change', update)
    onUnmounted(() => query.removeEventListener('change', update))

    return isNarrow
}
