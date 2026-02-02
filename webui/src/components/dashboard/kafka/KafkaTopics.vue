<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useMetrics } from '@/composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import Markdown from 'vue3-markdown-it'
import { getRouteName } from '@/composables/dashboard';
import { computed } from 'vue';
import { useLocalStorage } from '@/composables/local-storage';

const props = defineProps<{
    service: KafkaService,
}>()

const { sum } = useMetrics()
const { format } = usePrettyDates()
const router = useRouter()
const tags = useLocalStorage<string[]>(`kafka-${props.service.name}-tags`, ['__all'])

const allTags = computed(() => {
    if (!props.service || !props.service.topics) {
        return [];
    }
    const result: KafkaTag[] = [];
    const names: { [name: string]: any } = {}

    for (const topic of props.service.topics) {
        if (!topic.tags) {
            continue
        }
        for (const tag of topic.tags) {
            if (!names[tag.name]) {
                result.push(tag);
                names[tag.name] = tag;
            }
        }
    }
    return result;
})
const topics = computed(() => {
    if (!props.service.topics) {
        return []
    }
    let result = props.service.topics.sort(compareTopics)
    if (!tags.value.includes('__all')) {
        result = result.filter((t) => {
            if (t.tags && t.tags.some(tag => tags.value.some(x => x == tag.name))) {
                return true
            }
            return false;
        })
    }
    return result;
})

function compareTopics(t1: KafkaTopic, t2: KafkaTopic) {
    const name1 = t1.name.toLowerCase()
    const name2 = t2.name.toLowerCase()
    return name1.localeCompare(name2)
}

function messages(service: Service, topic: KafkaTopic) {
    return sum(service.metrics, 'kafka_messages_total{', { name: 'topic', value: topic.name })
}

function lastMessage(service: Service, topic: KafkaTopic) {
    const n = sum(service.metrics, 'kafka_message_timestamp{', { name: 'topic', value: topic.name })
    if (n == 0) {
        return '-'
    }
    return format(n)
}
function goToTopic(topic: KafkaTopic, openInNewTab = false) {
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
function toggleTag(name: string) {
    if (name === '__all') {
        if (tags.value.includes('__all')) {
            tags.value = []
        } else {
            tags.value = ['__all']
        }
    } else {
        if (tags.value.includes('__all')) {
            tags.value = allTags.value.map(x => x.name)
        }

        const i = tags.value.indexOf(name)
        if (i === -1) {
            tags.value.push(name)
        } else {
            tags.value.splice(i, 1)
        }

        tags.value = tags.value.filter((t) => t !== '__all')
    }
}
</script>

<template>
    <section class="card" aria-labelledby="topics">
        <div class="card-body">
            <h2 id="topics" class="card-title text-center">Topics</h2>

            <fieldset class="text-center mt-3 mb-2" v-if="allTags.length > 1" aria-describedby="tag-help">

                <legend class="visually-hidden">
                    Filter topics by tags
                </legend>

                <p id="tag-help" class="visually-hidden">
                    Select one or more tags to filter the topics. Selecting “All” enables all tags.
                </p>

                <div class="form-check form-check-inline">
                    <input class="form-check-input tag-checkbox" type="checkbox" id="all" value="all"
                        @change="toggleTag('__all')" :checked="tags.includes('__all')" aria-controls="tag-list">
                    <label class="form-check-label" for="all">All</label>
                </div>

                <div id="tag-list" style="display: inline-block">
                    <div class="form-check form-check-inline" v-for="tag in allTags" :key="tag.name" :title="tag.description">
                        <input class="form-check-input tag-checkbox" type="checkbox" :id="tag.name" :value="tag.name"
                            @change="toggleTag(tag.name)" :checked="tags.includes(tag.name) || tags.includes('__all')"
                            :aria-describedby="tag.description ? `desc-${tag.name}` : undefined">
                        <label class="form-check-label" :for="tag.name">{{ tag.name }}</label>

                        <span v-if="tag.description" :id="`desc-${tag.name}`" class="visually-hidden">
                            {{ tag.description }}
                        </span>
                    </div>
                </div>

            </fieldset>
            <div class="table-responsive-sm">
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
                        <tr v-for="topic in topics" :key="topic.name" @mouseup.left="goToTopic(topic)"
                            @mousedown.middle="goToTopic(topic, true)">
                            <td>
                                <router-link @click.stop class="row-link"
                                    :to="{ name: getRouteName('kafkaTopic').value, params: { service: props.service.name, topic: topic.name } }">
                                    {{ topic.name }}
                                </router-link>
                            </td>
                            <td>
                                <markdown :source="topic.description" class="description" :html="true"></markdown>
                            </td>
                            <td class="text-center">{{ lastMessage(service, topic) }}</td>
                            <td class="text-center">{{ messages(service, topic) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>