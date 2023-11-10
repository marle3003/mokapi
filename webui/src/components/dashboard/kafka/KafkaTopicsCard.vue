<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import type { PropType } from 'vue';
import Markdown from 'vue3-markdown-it'

const props = defineProps({
    service: { type: Object as PropType<KafkaService>, required: true },
})

const {sum} = useMetrics()
const {format} = usePrettyDates()
const router = useRouter()
const route = useRoute()

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
function goToTopic(topic: KafkaTopic){
    router.push({
        name: 'kafkaTopic',
        params: {
          service: props.service.name,
          topic: topic.name,
        },
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Topics</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Name</th>
                        <th scope="col" class="text-left">Description</th>
                        <th scope="col" class="text-center">Last Message</th>
                        <th scope="col" class="text-center">Messages</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="topic in service.topics" :key="topic.name" @click="goToTopic(topic)">
                        <td>{{ topic.name }}</td>
                        <td><markdown :source="topic.description" class="description"></markdown></td>
                        <td class="text-center">{{ lastMessage(service, topic) }}</td>
                        <td class="text-center">{{ messages(service, topic) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>