import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export function useExample() {
    const { formatLanguage } = usePrettyLanguage()

    function fetchExample(schema: Schema, contentType?: string) {
        if (!contentType) {
            contentType = 'application/json'
        }
        
        const response = useFetch('/api/schema/example', {method: 'POST', body: JSON.stringify(schema), headers: {'Content-Type': 'application/json', 'Accept': contentType}}, false, false)
        const example = ref<string>()

        watchEffect(() => {
            if (response.isLoading) {
                return
            }
            if (typeof response.data === 'string') {
                example.value = response.data
            } else {
                example.value = JSON.stringify(response.data)
            }
            example.value = formatLanguage(example.value, contentType!)
        })
        return example
    }

    return { fetchExample }
}