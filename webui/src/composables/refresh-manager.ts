import { computed, onUnmounted, ref } from "vue";
import router from '@/router';

const tasks = ref(new Map<string, () => void>());
const progress = ref(0)
const interval = computed(() => {
    const route = router.currentRoute.value
    return Number(route.query.refresh) * 1000
}) 
let startTime = Date.now(), timer = ref<number | undefined>();
const isActive = computed(() => timer !== undefined)

export function useRefreshManager() {

    function add(id: string, callback: () => void) {
        tasks.value.set(id, callback);
    }

    function remove(id: string) {
        tasks.value.delete(id);
    }

    function tick() {
        const now = Date.now();
        const elapsed = now - startTime;
        progress.value = Math.min((elapsed / interval.value) * 100, 100);

        if (elapsed >= interval.value) {
            for (const [_, callback] of tasks.value.entries()) {
                callback();
            }
            startTime = now;
            progress.value = 0;
        }
    }

    function start() {
        if (!timer.value) {
            timer.value = setInterval(tick, 100);
        }
    }

    function stop() {
        clearInterval(timer.value);
        timer.value = undefined;
    }

    onUnmounted(stop)

    return { add, remove, start, stop, progress, isActive }
}