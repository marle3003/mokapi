import { onMounted, onUnmounted } from "vue";

export function useShortcut(key: string, handler: (e: KeyboardEvent) => void) {
    onMounted(() => window.addEventListener('keyup', handler))
    onUnmounted(() => window.removeEventListener('keyup', handler))
}