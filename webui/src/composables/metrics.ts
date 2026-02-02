export function useMetrics() {

    function max(metrics: Metric[], name: string, ...labels: Label[]): number {
        if (!metrics){
            return 0
        }

        let max = 0
        for (let metric of metrics) {
            if (!metric.name.startsWith(name)) {
                continue
            }
            
            if (labels.length == 0 || matchLabels(metric, labels)){
                const n = Number(metric.value)
                if (n > max) {
                    max = n
                }
            }
        }  
        return max
    }

    function sum(metrics: Metric[], name: string, ...labels: Label[]): number{
        if (!metrics){
            return 0
        }
    
        let sum = 0
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

    function value(metrics: Metric[], name: string, ...labels: Label[]): number | undefined {
        if (!metrics){
            return 0
        }

        for (let metric of metrics) {
            if (!metric.name.startsWith(name)) {
                continue
            }
            
            if (labels.length == 0 || matchLabels(metric, labels)){
                return Number(metric.value)
            }
        }  
        return undefined
    }

    function filter(metrics: Metric[], name: string, ...labels: Label[]): Metric[] {
        const result = [];
        if (!metrics){
            return [];
        }

        for (let metric of metrics) {
            if (!metric.name.startsWith(name)) {
                continue
            }
            
            if (labels.length == 0 || matchLabels(metric, labels)){
                result.push(metric)
            }
        }  
        return result;
    }

    function parseLabels(metric: Metric) {
        const match = metric.name.match(/\{(.+)}/);
        if (!match || match.length < 2 || !match[1]) {
            return {};
        }

        return Object.fromEntries(
            Array.from(
                match[1].matchAll(/(\w+)="([^"]*)"/g),
                m => [m[1], m[2]]
            )
        );
    }

    return {sum, max, value, filter, parseLabels}
}