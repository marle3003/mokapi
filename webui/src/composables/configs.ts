import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'

export function useConfig() {

    function fetch(id: string) {
        const response = useFetch(`/api/configs/${id}`)
        const config = ref<Config | null>(null)
        const isLoading = ref<boolean>()

        watchEffect(() => {
            config.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {config, isLoading, close: response.close}
    }

    function fetchData(id: string) {
        const response = useFetch(getDataUrl(id))
        const data = ref<string | null>(null)
        const isLoading = ref<boolean>()

        watchEffect(() => {
            data.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {data, isLoading, close: response.close}
    }

    function getDataUrl(id: string): string {
        return `/api/configs/${id}/data`
    }

    return { fetch, fetchData, getDataUrl }
}