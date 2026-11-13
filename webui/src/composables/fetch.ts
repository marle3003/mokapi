import { reactive } from 'vue'
import { useRefreshManager } from './refresh-manager';
import { transformPath } from './site-base'
import { useConnection } from './connection'

export interface Response {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- response body and error depend on the endpoint, callers narrow them
    data: any
    isLoading: boolean
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- response body and error depend on the endpoint, callers narrow them
    error: any
    close: () => void
    refs: number
    refresh: () => void
    header?: Headers
}

const cache: { [name: string]: Response } = {}
const manager = useRefreshManager();
const connection = useConnection();
// Fetches without refresh are not reloaded by the refresh manager, so the ones that never reached
// Mokapi are repeated once after reconnect.
const failedOneOffFetches = new Set<() => void>()
connection.onReconnect(manager.refreshAll)
connection.onReconnect(() => {
    const fetches = [...failedOneOffFetches]
    failedOneOffFetches.clear()
    fetches.forEach(doFetch => doFetch())
})

export function useFetch(path: string, options?: RequestInit, doRefresh: boolean = true, useCache: boolean = true): Response {
    path = transformPath(path)
    const cached = cache[path]
    const response: Response = cached || reactive({
        data: null,
        isLoading: false,
        error: null,
        close: () => {},
        refs: 1,
        refresh: doFetch,
    })

    if (cache[path]) {
        response.refs++
        return response
    }

    if (useCache) {
        cache[path] = response
    }

    function doFetch() {
        failedOneOffFetches.delete(doFetch)
        response.isLoading = true
        connection.monitoredFetch(path, options).then(readResult, handleNetworkError)
    }

    // While the connection is monitored, keep data and stay loading until it is back.
    function handleNetworkError(err: Error) {
        if (!connection.isMonitoring()) {
            handleError(err)
            return
        }
        if (!doRefresh) {
            failedOneOffFetches.add(doFetch)
        }
    }

    function readResult(res: globalThis.Response) {
        return readBody(res)
            .then((body) => {
                response.data = body
                response.isLoading = false
            })
            .catch(handleError)
    }

    function handleError(err: Error) {
        let msg = err.toString()
        if (!msg) {
            msg = 'Network connection error'
        }
        console.error(err)
        response.error = msg
        response.data = null
        response.isLoading = false
    }

    async function readBody(res: globalThis.Response) {
        if (!res.ok) {
            const text = await res.text()
            throw new Error(res.statusText + ': ' + text)
        }
        const contentType = res.headers.get("content-type");
        response.header = res.headers
        if (contentType && contentType.indexOf("application/json") !== -1) {
            return res.json()
        } else{
            return res.text()
        }
    }

    if (doRefresh){
        manager.add(path, doFetch)
        response.close = function() {
            response.refs--
            if (response.refs == 0) {
                manager.remove(path)
                delete cache[path]
            }
        }
    }
    doFetch()
    
    return response
}
