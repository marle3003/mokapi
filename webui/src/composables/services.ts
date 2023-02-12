import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'

export function useService() {
    
    function fetchServices(type?: string) {
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

    function fetchService(name: string, type: string) {
        const response = useFetch(`/api/services/${type}/${name}`)
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