import { computed, onUnmounted, ref } from "vue";
import router from '@/router';

const tasks = ref(new Map<string, () => void>());
const progress = ref(0)
const interval = computed(() => {
    const route = router.currentRoute.value
    return Number(route.query.refresh) * 1000
}) 
let startTime = Date.now()
const timer = ref<ReturnType<typeof setInterval> | undefined>();
const isActive = computed(() => 
    timer.value !== undefined
)
let isPaused = false

export function useRefreshManager() {

    function add(id: string, callback: () => void) {
        tasks.value.set(id, callback);
    }

    function remove(id: string) {
        tasks.value.delete(id);
    }

    function refreshAll() {
        for (const callback of tasks.value.values()) {
            callback();
        }
    }

    function tick() {
        const now = Date.now();
        const elapsed = now - startTime;
        progress.value = Math.min((elapsed / interval.value) * 100, 100);

        if (elapsed >= interval.value) {
            refreshAll();
            startTime = now;
            progress.value = 0;
        }
    }

    function startTimer() {
        startTime = Date.now();
        progress.value = 0;
        timer.value = setInterval(tick, 100);
    }

    function start() {
        if (!timer.value) {
            timer.value = setInterval(tick, 100);
        }
        onUnmounted(stop)
    }

    function stop() {
        clearInterval(timer.value);
        timer.value = undefined;
        isPaused = false;
    }

    // Pausing hides the progress bar and stops the refresh cycle, resuming restarts it at 0.
    // Without a running cycle nothing is paused or resumed.
    function pause() {
        if (timer.value) {
            stop();
            isPaused = true;
        }
    }

    function resume() {
        if (isPaused) {
            isPaused = false;
            startTimer();
        }
    }

    return { add, remove, refreshAll, start, stop, pause, resume, progress, isActive }
}