import { useFetch } from './fetch'
import { onUnmounted } from 'vue'

export function useMetrics() {
    let responses: Response[] = []

    function query(query: string): Response{
        const response = useFetch('/api/metrics?q=' + query)
        responses.push(response)
        return response
    }

    function sum(metrics: Metric[], name: string, ...labels: Label[]): number{
        let sum = 0
        if (!metrics){
            return 0
        }
    
        for (let metric of metrics) {
            if (!metric.name.startsWith(name)) {
                continue
            }
            
            if (labels.length == 0 || matchLabels(metric, labels)){
                sum += Number(metric.value)
            }
        }  
        return sum
    }

    const matchLabels = (metric: Metric, labels: Label[]): Boolean => {
        for (var label of labels){
            const s = `${label.name}="${label.value}"`
            if (!metric.name.includes(s)){
                return false
            }
        }
        return true
    }

    return {query, sum}
}