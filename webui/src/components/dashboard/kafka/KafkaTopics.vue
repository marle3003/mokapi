<script setup lang="ts">
import PinButton from '@/components/dashboard/layout/PinButton.vue'
import { useRouter } from 'vue-router'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard';
import { computed } from 'vue';
import { useLocalStorage } from '@/composables/local-storage';
import { splitTemplate, instanceCount } from '@/composables/channel'
import { groupRows, selectedTags, tagNames, type TagGrouping } from '@/composables/tag-group'
import TagToolbar from '@/components/dashboard/TagToolbar.vue'
import TagGroupHeading from '@/components/dashboard/TagGroupHeading.vue'

const props = defineProps<{
    service: KafkaService,
}>()

const { format } = usePrettyDates()
const router = useRouter()
const tags = useLocalStorage<string[]>(`kafka-${props.service.name}-tags`, ['__all'])
const grouping = useLocalStorage<TagGrouping>(`kafka-${props.service.name}-group`, 'az')

const allTags = computed(() => {
    if (!props.service || !props.service.topics) {
        return [];
    }
    const result: Tag[] = [];
    const names: { [name: string]: Tag } = {}

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
const filteredTopics = computed(() => {
    if (!props.service.topics) {
        return []
    }
    let result = [...props.service.topics]
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

const groups = computed(() => groupRows(filteredTopics.value, grouping.value, {
    nameOf: topic => topic.name,
    tagsOf: tagNames,
    tags: allTags.value,
    onlyTags: selectedTags(tags.value),
}))

function lastMessage(topic: KafkaTopicInfo) {
    if (!topic.metrics.message_timestamp || topic.metrics.message_timestamp == 0) {
        return '-'
    }
    return format(topic.metrics.message_timestamp)
}
function goToTopic(topic: KafkaTopicInfo, openInNewTab = false) {
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
function instanceLabel(channel: KafkaTopicInfo) {
  const n = instanceCount(channel)
  if (n === 0) return 'no topics yet'
  return n === 1 ? '1 topic' : `${n} topics`
}
</script>

<template>
    <section class="card" aria-labelledby="topics">
        <div class="card-body">
            <h2 id="topics" class="card-title text-center">Topics</h2>

            <tag-toolbar v-model:selected="tags" v-model:grouping="grouping" :tags="allTags" noun="topics" />
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" aria-labelledby="topics">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-4">Name</th>
                            <th scope="col" class="text-left col-4">Summary</th>
                            <th scope="col" class="text-center col-2">Last Message</th>
                            <th scope="col" class="text-center col-1">Messages</th>
                        </tr>
                    </thead>
                    <tbody v-for="group in groups" :key="group.heading ?? ''">
                        <tag-group-heading v-if="group.heading" :heading="group.heading" :description="group.description" :columns="4" />
                        <tr v-for="topic in group.rows" :key="topic.name" @mouseup.left="goToTopic(topic)"
                            @mousedown.middle="goToTopic(topic, true)">
                            <td>
                                <pin-button class="float-end" :to="{ name: getRouteName('kafkaTopic').value, params: { service: props.service.name, topic: topic.name } }" :label="topic.name" />
                                <router-link @click.stop class="row-link"
                                    :to="{ name: getRouteName('kafkaTopic').value, params: { service: props.service.name, topic: topic.name } }">
                                    <template v-if="topic.parameters?.length">
                                        <span class="visually-hidden">Template </span>
                                        <span class="channel-name">
                                            <template v-for="(part, i) in splitTemplate(topic.name)" :key="i">
                                                <span v-if="part.param" class="param">{{ part.text }}</span>
                                                <template v-else>{{ part.text }}</template>
                                            </template>
                                        </span>
                                        <span class="visually-hidden">, </span>
                                        <span class="badge ms-2"
                                        :class="instanceCount(topic) ? 'bg-secondary' : 'bg-light text-muted'">
                                            {{ instanceLabel(topic) }}
                                        </span>
                                    </template>
                                    <span v-else>{{ topic.name }}</span>
                                </router-link>
                            </td>
                            <td>
                                <div class="table-markdown">{{ topic.summary }}</div>
                            </td>
                            <td class="text-center">{{ lastMessage(topic) }}</td>
                            <td class="text-center">{{ topic.metrics.messages_total }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>