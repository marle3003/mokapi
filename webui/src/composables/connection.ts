import { readonly, ref } from 'vue'
import { transformPath } from './site-base'

const heartbeatInterval = 5000

const connected = ref(true)
const reconnectCallbacks: (() => void)[] = []
let timer: ReturnType<typeof setInterval> | undefined

// Only the live dashboard starts the monitor. Without it network errors are ignored, e.g. of the
// static data in the demo dashboard.
function start() {
    if (timer === undefined) {
        timer = setInterval(heartbeat, heartbeatInterval)
    }
}

function stop() {
    clearInterval(timer)
    timer = undefined
    connected.value = true
}

function isMonitoring() {
    return timer !== undefined
}

function reportNetworkError() {
    if (isMonitoring()) {
        connected.value = false
    }
}

// A rejected fetch never reached Mokapi. HTTP errors are responses and don't count.
function monitoredFetch(input: RequestInfo | URL, init?: RequestInit) {
    return fetch(input, init).catch((err) => {
        reportNetworkError()
        throw err
    })
}

function onReconnect(callback: () => void) {
    reconnectCallbacks.push(callback)
}

// Any HTTP response, also 4xx and 5xx, means that Mokapi is reachable.
async function heartbeat() {
    try {
        await monitoredFetch(transformPath('/api/info'))
    } catch {
        return
    }
    if (!connected.value) {
        connected.value = true
        reconnectCallbacks.forEach(callback => callback())
    }
}

export function useConnection() {
    return { isConnected: readonly(connected), start, stop, isMonitoring, monitoredFetch, onReconnect }
}
