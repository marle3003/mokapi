<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { useMetrics } from '@/composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import Markdown from 'vue3-markdown-it'
import { getRouteName } from '@/composables/dashboard';

const props = defineProps<{
    service: KafkaService,
}>()

const { sum } = useMetrics()
const { format } = usePrettyDates()
const router = useRouter()

function messages(service: Service, topic: KafkaTopic){
    return sum(service.metrics, 'kafka_messages_total{', {name: 'topic', value: topic.name})
}

function lastMessage(service: Service, topic: KafkaTopic){
    const n = sum(service.metrics, 'kafka_message_timestamp{', {name: 'topic', value: topic.name})
    if (n == 0){
        return '-'
    }
    return format(n)
}
function goToTopic(topic: KafkaTopic, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaTopic').value,
        params: {
          service: props.service.name,
          topic: topic.name,
        }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
</script>

<template>
    <section class="card" aria-labelledby="topics">
        <div class="card-body">
            <h2 id="topics" class="card-title text-center">Topics</h2>
            <table class="table dataTable selectable" aria-labelledby="topics">
                <thead>
                    <tr>
                        <th scope="col" class="text-left col-4">Name</th>
                        <th scope="col" class="text-left col-4">Description</th>
                        <th scope="col" class="text-center col-2">Last Message</th>
                        <th scope="col" class="text-center col-1">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="topic in service.topics" :key="topic.name" @mouseup.left="goToTopic(topic)" @mousedown.middle="goToTopic(topic, true)">
                        <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('kafkaTopic').value, params: { service: props.service.name, topic: topic.name }}">
                                {{ topic.name }}
                            </router-link>
                        </td>
                        <td><markdown :source="topic.description" class="description" :html="true"></markdown></td>
                        <td class="text-center">{{ lastMessage(service, topic) }}</td>
                        <td class="text-center">{{ messages(service, topic) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>