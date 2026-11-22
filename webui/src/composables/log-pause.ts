import { ref } from 'vue'

// Pause of one log. It lives in memory only and is not shared: another log, a reload or a
// new tab starts running again.
export function useLogPause() {
    const paused = ref(false)

    function toggle() {
        paused.value = !paused.value
    }

    function resume() {
        paused.value = false
    }

    return { paused, toggle, resume }
}
