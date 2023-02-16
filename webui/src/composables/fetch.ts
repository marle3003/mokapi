import { onUnmounted, reactive } from 'vue'
import { useBackendBaseUrl } from './backendBaseUrl'
import router from '@/router';

let cache: {[name: string]: any} = {}

export function useFetch(path: string, options?: RequestInit, doRefresh: boolean = true) {
    const baseUrl = useBackendBaseUrl()
    const route = router.currentRoute.value
    const cached = cache[path]
    const response = cached || reactive({
        data: null,
        isLoading: false,
        error: null,
        stop: null
    })

    if (cache[path]) {
        return response
    }
    cache[path] = response

    function doFetch() {
        response.isLoading = true
        fetch(baseUrl + path, options)
            .then((res) => res.json())
            .then((res) => response.data = res)
            .catch((err) => {
                response.error = 'Network connection error'
                response.data = null
            })
            .then(() => response.isLoading = false)
    }

    const refresh = Number(route.query.refresh)
    if (refresh && doRefresh){
        const timer = setInterval(doFetch, refresh * 1000)
        onUnmounted(() => {
            clearInterval(timer)
        })
        response.stop = function() {
            clearInterval(timer)
        }
    }
    doFetch()
    
    return response
}