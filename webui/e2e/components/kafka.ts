import { Page } from "playwright/test";

export function useKafka(page: Page) {
    return {
        metrics: useMetrics(page),
        clusters: page.getByRole('region', { name: 'Kafka Clusters' })
    }
}

export function useMetrics(page: Page) {
    return {
        messages: page.getByRole('marquee', { name: 'Kafka Messages' })
    }
}