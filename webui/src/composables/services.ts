import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'
import type {Metric} from './metrics'

export interface Service {
    name: string
    description: string
    version: string
    contact: Contact | null
    type: ServiceType
    metrics: Metric[]
}

export interface HttpService extends Service {
    paths: HttpPath[]
}

export interface HttpPath {
    path: string
    summary: string
    description: string
    operations: HttpOperation[]
}

export interface HttpOperation {
    method: string
}

export interface Contact {
    name: string
    url: string
    email: string
}

export enum ServiceType{
    Http = "http",
    Kafka = "kafka",
    Smtp = "smtp"
}

export function useService() {
    
    function fetchServices(type?: ServiceType) {
        const response = useFetch('/api/services')
        const services = ref<Service[]>([])

        watchEffect(() => {
            if (!response.data){
                services.value = []
                return
            }
    
            let result = []
            if (type) {
                for (let service of response.data) {
                    if (service.type == type){
                        result.push(service)
                    }
                }
            }else{
                result = response.data
            }
    
            services.value = result.sort(compareService)
        })
        return services
    }

    function fetchService(name: string) {
        const response = useFetch('/api/services/http/' + name)
        const service = ref<Service | null>(null)

        watchEffect(() => {
            service.value = response.data ? response.data : null
        })
        return service
    }

    return {fetchServices, fetchService}
}

function compareService(s1: Service, s2: Service) {
    const name1 = s1.name.toLowerCase()
    const name2 = s2.name.toLowerCase()
    return name1.localeCompare(name2)
}