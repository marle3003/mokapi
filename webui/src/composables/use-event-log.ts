import { computed, reactive, ref, watch } from 'vue'
import { useRoute, type RouteLocationNormalizedLoaded } from 'vue-router'
import { useDashboard } from '@/composables/dashboard'
import { useLogPause } from '@/composables/log-pause'
import { usePageSize } from '@/composables/paging'
import type { HttpEventFilter } from '@/types/dashboard'

type LogQuery = { labels: Label[], http?: HttpEventFilter }

// The log only reloads while the auto refresh runs, which needs a refresh interval in the route
// (?refresh=<seconds>, see refresh-manager.ts). Without it, a pause has nothing to stop.
function autoRefreshIsOn(route: RouteLocationNormalizedLoaded) {
    const seconds = Number(route.query.refresh)
    return Number.isFinite(seconds) && seconds > 0
}

// The recorded events of a log (requests, messages): the current page, reloaded while the log is
// shown, and the Pause that stops the reload. `query` is what the log asks for besides the page.
export function useEventLog(askedFor: () => LogQuery) {
    // Compared by content: a log that is built from reloaded data (the service of its view) rebuilds
    // an equal query on every reload, which must neither end the pause nor restart the request. The
    // text of the last query is kept, because the query itself may share objects that are edited in place.
    let lastText = ''
    let last: LogQuery | undefined
    const query = computed<LogQuery>(() => {
        const next = askedFor()
        const text = JSON.stringify(next)
        if (!last || text !== lastText) {
            last = next
            lastText = text
        }
        return last
    })
    const paging = reactive({
        current: 1,
        size: usePageSize(25),
    })

    const { dashboard, getMode } = useDashboard()
    const route = useRoute()
    // The demo has no live events, so there is nothing to pause.
    const canPause = computed(() => getMode() === 'live' && autoRefreshIsOn(route))
    const { paused, toggle: togglePause, resume } = useLogPause()
    const result = ref<ServiceEventPage | null>(null)
    const isLoading = ref(true)
    const events = computed(() => result.value?.items ?? [])

    // A new query or page asks for other events, which ends the pause.
    watch([() => paging.current, () => paging.size, query], resume)
    watch(
        [() => dashboard.value, () => paging.current, () => paging.size, query, paused],
        ([db], _, onCleanup) => {
            // closing the previous fetch (onCleanup) stops its reload, the last result stays on the screen
            if (paused.value) {
                return
            }
            const res = db.getEvents({ page: paging.current, size: paging.size, ...query.value })

            const stop = watch([res.data, res.isLoading], ([v, loading]) => { result.value = v; isLoading.value = loading }, { immediate: true })

            onCleanup(() => {
                stop()
                res.close()
            })
        },
        { immediate: true }
    )

    return { paging, result, isLoading, events, paused, canPause, togglePause }
}
