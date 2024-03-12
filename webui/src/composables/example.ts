import { watchEffect, reactive } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export interface Example {
    data: string
    error: string | null
    next:  () => void
}

export function useExample() {
    const { formatLanguage } = usePrettyLanguage()

    function fetchExample(schema: Schema, contentType?: string) {
        if (!contentType) {
            contentType = 'application/json'
        }
        
        const response = useFetch('/api/schema/example', {method: 'POST', body: JSON.stringify(schema), headers: {'Content-Type': 'application/json', 'Accept': contentType}}, false, false)
        const example: Example =  reactive({
            data: '',
            next: () => response.refresh(),
            error: null
        })

        watchEffect(() => {
            if (response.isLoading) {
                return
            }
            if (response.error) {
                example.error = response.error
                example.data = ''
            }

            let data = ''
            if (typeof response.data === 'string') {
                data = response.data
            } else {
                data = JSON.stringify(response.data)
            }
            example.data = formatLanguage(data, contentType!)
        })
        return example
    }

    return { fetchExample }
}