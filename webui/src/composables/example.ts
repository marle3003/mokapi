import { watchEffect, ref } from 'vue'
import { useFetch } from './fetch'
import { usePrettyLanguage } from './usePrettyLanguage'

export function useExample() {
    const {formatLanguage} = usePrettyLanguage()
    
    function fetchExample(schema: Schema) {
        const response = useFetch('/api/schema/example', {method: 'POST', body: JSON.stringify(schema), headers: {'Content-Type': 'application/json', 'Accept': 'application/json'}}, false, false)
        const example = ref<string>()

        watchEffect(() => {
                example.value = formatLanguage(JSON.stringify(response.data), 'application/json')
        })
        return example
    }

    return {fetchExample}
}