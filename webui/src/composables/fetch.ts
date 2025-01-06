import { reactive } from 'vue'
import router from '@/router';

export interface Response {
    data: any
    isLoading: boolean
    error: any
    close: () => void
    refs: number
    refresh: () => void
    header?: Headers
}

const cache: { [name: string]: Response } = {}

export function useFetch(path: string, options?: RequestInit, doRefresh: boolean = true, useCache: boolean = true): Response {
    path = transformPath(path)
    const route = router.currentRoute.value
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
        response.isLoading = true
        fetch(path, options)
            .then(async (res) => {
                if (!res.ok) {
                    let text = await res.text()
                    throw new Error(res.statusText + ': ' + text)
                }
                const contentType = res.headers.get("content-type");
                response.header = res.headers
                if (contentType && contentType.indexOf("application/json") !== -1) {
                    return res.json()
                } else{
                    return res.text()
                }
            })
            .then((res) => {
                response.data = res
                response.isLoading = false
            })
            .catch((err) => {
                let msg = err.toString()
                if (!msg) {
                    msg = 'Network connection error'
                }
                console.error(err)
                response.error = msg
                response.data = null
                response.isLoading = false
            })
    }

    const refresh = Number(route.query.refresh)
    if (refresh && doRefresh){
        const timer = setInterval(doFetch, refresh * 1000)
        response.close = function() {
            response.refs--
            if (response.refs == 0) {
             clearInterval(timer)
             delete cache[path]
            }
        }
    }
    doFetch()
    
    return response
}

export function transformPath(path: string): string {
    let base = document.querySelector('base')?.href
    if (base) {
        base = base.substring(0, base.length - 1)
        path = base + path
    }
    return path
}