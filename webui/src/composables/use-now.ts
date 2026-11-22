import { onUnmounted, ref } from 'vue'

// The current time in milliseconds, updated every `intervalMs` while the component is mounted.
export function useNow(intervalMs = 1000) {
    const now = ref(Date.now())
    const interval = setInterval(() => { now.value = Date.now() }, intervalMs)
    onUnmounted(() => clearInterval(interval))
    return now
}
