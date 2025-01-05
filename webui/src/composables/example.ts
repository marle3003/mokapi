import { watchEffect, reactive } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export interface Example {
    data: string
    error: string | null
    next:  () => void
}

export interface Request {
    name?: string
    schema: SchemaFormat
    contentType?: string
}

export function useExample() {
    const { formatLanguage } = usePrettyLanguage()

    function fetchExample(request: Request) {
        if (!request.contentType) {
            request.contentType = 'application/json'
        }
        
        const response = useFetch('/api/schema/example', {
            method: 'POST', 
            body: JSON.stringify({name: request.name, schema: request.schema.schema, format: request.schema.format}), 
            headers: {'Content-Type': 'application/json', 'Accept': request.contentType}},
            false, false)
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
            example.data = formatLanguage(data, request.contentType!)
        })
        return example
    }

    return { fetchExample }
}