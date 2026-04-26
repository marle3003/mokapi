import { ref } from 'vue'

export function useProgressiveLoading(options = {}) {
    const {
        steps = [
            { delay: 0, text: 'Searching…' },
            { delay: 3000, text: 'Still searching…' },
            { delay: 8000, text: 'Taking longer than usual…' }
        ]
    }: { steps?: { delay: number, text: string }[] } = options

    const isLoading = ref(false)
    const statusText = ref('')

    let timers: number[] = []
    let currentId = 0

    function start() {
        clear()

        isLoading.value = true
        const id = ++currentId

        steps.forEach(step => {
            const t = setTimeout(() => {
                if (id !== currentId) return
                statusText.value = step.text
            }, step.delay)

            timers.push(t)
        })


    }

    function stop() {
        clear()
        isLoading.value = false
        statusText.value = ''
    }

    function clear() {
        timers.forEach(t => clearTimeout(t))
        timers = []
    }

    return {
        isLoading,
        statusText,
        start,
        stop
    }
}
