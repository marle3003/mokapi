import { watchEffect, reactive } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export interface Response {
    data: Example[]
    error: string | null
    next:  () => void
}

export interface Example {
    contentType: string
    value: string
    error?: string
}

export interface Request {
    name?: string
    schema: SchemaFormat
    contentTypes?: string[]
}

export function useExample() {
    const { formatLanguage } = usePrettyLanguage()

    function fetchExample(request: Request) {
      
        const response = useFetch('/api/schema/example', {
            method: 'POST', 
            body: JSON.stringify({name: request.name, schema: request.schema.schema, format: request.schema.format, contentTypes: request.contentTypes}), 
            headers: {'Content-Type': 'application/json', 'Accept': 'application/json'}},
            false, false)
        const res: Response =  reactive({
            data: [],
            next: () => response.refresh(),
            error: null
        })

        watchEffect(() => {
            if (response.isLoading) {
                return
            }
            if (response.error) {
                res.error = response.error
                res.data = []
                return
            }

            res.data = response.data
            for (const example of res.data) {
                example.value = atob(example.value)
                example.value = formatLanguage(example.value, example.contentType!)
            }
        })
        return res
    }

    return { fetchExample }
}