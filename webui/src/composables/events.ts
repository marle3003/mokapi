import { ref, watchEffect } from 'vue'
import { useFetch } from './fetch'

export function useEvents() {

    function fetch(namespace: string, ...labels: Label[]){
        let url = `/api/events?namespace=${namespace}`
        for (let label of labels) {
            url += `&${label.name}=${label.value}`
        }
        const res = useFetch(url)
        const events = ref<ServiceEvent[]>()

        watchEffect(() => {
            if (!res.data){
                events.value = []
                return
            }
            events.value = res.data
        })
        return { events, close: res.close }
    }

    function fetchById(id: string){
        const event = ref<ServiceEvent>()
        const isLoading = ref<Boolean>()
        const response = useFetch(`/api/events/${id}`)
        watchEffect(() =>{
            event.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {event, isLoading, close: response.close}
    }

    return {fetch, fetchById}
}