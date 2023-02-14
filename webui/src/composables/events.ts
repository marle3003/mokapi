import { onUnmounted, ref, watchEffect } from 'vue'
import { useFetch } from './fetch'

interface Response{
    stop: () => void
}

export function useEvents() {
    let responses: Response[] = []

    function fetch(namespace: string, ...labels: Label[]){
        let url = `/api/events?namespace=${namespace}`
        for (let label of labels) {
            url += `&${label.name}=${label.value}`
        }
        const events = ref<ServiceEvent[]>()
        const isLoading = ref<Boolean>()
        const response = useFetch(url)
        responses.push(response)
        watchEffect(() =>{
            events.value = response.data ? response.data : []
            isLoading.value = !response.data && response.isLoading
        })
        return {events, isLoading}
    }

    function fetchById(id: string){
        const event = ref<ServiceEvent>()
        const isLoading = ref<Boolean>()
        const response = useFetch(`/api/events/${id}`)
        watchEffect(() =>{
            event.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {event, isLoading}
    }

    onUnmounted(() => {
        for (let response of responses){
            if (response.stop){
                response.stop()
            }
        }
    })

    return {fetch, fetchById}
}