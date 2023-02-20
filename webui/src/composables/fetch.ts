import { reactive } from 'vue'
import { useBackendBaseUrl } from './backendBaseUrl'
import router from '@/router';

let cache: {[name: string]: any} = {}

export function useFetch(path: string, options?: RequestInit, doRefresh: boolean = true, useCache: boolean = true): Response {
    const baseUrl = useBackendBaseUrl()
    const route = router.currentRoute.value
    const cached = cache[path]
    const response = cached || reactive({
        data: null,
        isLoading: false,
        error: null,
        close: function() {},
        refs: 1
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
        fetch(baseUrl + path, options)
            .then((res) => res.json())
            .then((res) => {
                response.data = res
                response.isLoading = false
            })
            .catch((err) => {
                response.error = 'Network connection error'
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