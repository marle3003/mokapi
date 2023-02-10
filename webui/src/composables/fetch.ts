import { reactive } from 'vue'
import { useBackendBaseUrl } from './backendBaseUrl'
import router from '@/router';

let cache: {[name: string]: any} = {}

export function useFetch(path: string) {
    const baseUrl = useBackendBaseUrl()
    const route = router.currentRoute.value
    const cached = cache[path]
    const response = cached || reactive({
        data: null,
        isLoading: false,
        error: null
    })

    if (cache[path]) {
        return response
    }
    cache[path] = response

    function doFetch() {
        response.isLoading = true
        fetch(baseUrl + path)
            .then((res) => res.json())
            .then((res) => response.data = res)
            .catch((err) => {
                response.error = err
                response.data = null
                console.log("error fetch "+path+": "+err)
            })
            .then(() => response.isLoading = false)
    }

    const refresh = Number(route.query.refresh)
    if (refresh){
        setInterval(doFetch, refresh * 1000)
    }
    doFetch()
    
    return response
}