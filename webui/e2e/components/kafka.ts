import { Page } from "playwright/test"
import { useTable } from '../components/table'

export function useKafkaOverview(page: Page) {
    return {
        metrics: useMetrics(page),
        clusters: async () => await useTable(page.getByRole('region', { name: 'Kafka Clusters' }).getByRole('table', { name: 'Kafka Clusters' }))
    }
}

export function useMetrics(page: Page) {
    return {
        messages: page.getByRole('status', { name: 'Kafka Messages' })
    }
}

export function useKafkaCluster(page: Page) {
    return {
        summary: page.getByRole('region', { name: "Summary" })
    }
}