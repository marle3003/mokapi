import { ref, watchEffect } from 'vue'
import { useFetch } from './fetch'

export interface Event {
    id: string
    data: HttpEventData
    time: number
}

export interface HttpEventData {
    request: HttpEventRequest
    response: HttpEventResponse
    duration: number
}

export interface HttpEventRequest {
    method: string
    url: string
    contentType: string
    parameters: HttpEventParameter[]
    body: string
}

export interface HttpEventParameter {
    name: string
    type: string
    value: string
    raw: string
}

export interface HttpEventResponse {
    statusCode: number
    body: string
    size: number
    headers: HttpHeader
}

export interface HttpHeader  {[name: string]: string}

export function useEvents() {

    function fetch(namespace: string, name?: string, path?: string){
        let url = `/api/events?namespace=${namespace}`
        if (name){
            url += `&name=${name}`
        }
        if (path){
            url += `&path=${path}`
        }
        const events = ref<Event[]>()
        const isLoading = ref<Boolean>()
        const response = useFetch(url)
        watchEffect(() =>{
            events.value = response.data ? response.data : []
            isLoading.value = !response.data && response.isLoading
        })
        return {events, isLoading}
    }

    function fetchById(id: number){
        const event = ref<Event>()
        const isLoading = ref<Boolean>()
        const response = useFetch(`/api/events/${id}`)
        watchEffect(() =>{
            event.value = response.data ? response.data : null
            isLoading.value = response.isLoading
        })
        return {event, isLoading}
    }

    return {fetch, fetchById}
}