import { watchEffect, reactive } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export function useExample() {
    const { formatLanguage } = usePrettyLanguage()

    function fetchExample(schema: Schema, contentType?: string) {
        if (!contentType) {
            contentType = 'application/json'
        }
        
        const response = useFetch('/api/schema/example', {method: 'POST', body: JSON.stringify(schema), headers: {'Content-Type': 'application/json', 'Accept': contentType}}, false, false)
        const example =  reactive({
            data: '',
            next: () => response.refresh()
        })

        watchEffect(() => {
            if (response.isLoading) {
                return
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