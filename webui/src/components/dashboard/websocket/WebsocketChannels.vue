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
    service: WebsocketService,
}>()

const { format } = usePrettyDates()
const router = useRouter()
const tags = useLocalStorage<string[]>(`websocket-${props.service.name}-tags`, ['__all'])
const grouping = useLocalStorage<TagGrouping>(`websocket-${props.service.name}-group`, 'az')

const allTags = computed(() => {
    if (!props.service || !props.service.channels) {
        return [];
    }
    const result: Tag[] = [];
    const names: { [name: string]: Tag } = {}

    for (const channel of props.service.channels) {
        if (!channel.tags) {
            continue
        }
        for (const tag of channel.tags) {
            if (!names[tag.name]) {
                result.push(tag);
                names[tag.name] = tag;
            }
        }
    }
    return result;
})
const filteredChannels = computed(() => {
    if (!props.service.channels) {
        return []
    }
    let result = [...props.service.channels]
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

const groups = computed(() => groupRows(filteredChannels.value, grouping.value, {
    nameOf: channel => channel.name,
    tagsOf: tagNames,
    tags: allTags.value,
    onlyTags: selectedTags(tags.value),
}))

function lastMessage(channel: WebsocketChannel) {
    if (!channel.metrics.message_timestamp || channel.metrics.message_timestamp == 0) {
        return '-'
    }
    return format(channel.metrics.message_timestamp)
}

function goToChannel(channel: WebsocketChannel, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('websocketChannel').value,
        params: {
            service: props.service.name,
            channel: channel.name,
        }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function instanceLabel(channel: WebsocketChannel) {
  const n = instanceCount(channel)
  if (n === 0) return 'no paths yet'
  return n === 1 ? '1 path' : `${n} paths`
}
</script>

<template>
    <section class="card" aria-labelledby="channels">
        <div class="card-body">
            <h2 id="channels" class="card-title text-center">Paths</h2>

            <tag-toolbar v-model:selected="tags" v-model:grouping="grouping" :tags="allTags" noun="channels" />
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" aria-labelledby="channels">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-4">Path</th>
                            <th scope="col" class="text-left col-4">Summary</th>
                            <th scope="col" class="text-center col-2">Last Message</th>
                            <th scope="col" class="text-center col-1">Messages</th>
                        </tr>
                    </thead>
                    <tbody v-for="group in groups" :key="group.heading ?? ''">
                        <tag-group-heading v-if="group.heading" :heading="group.heading" :description="group.description" :columns="4" />
                        <tr v-for="channel in group.rows" :key="channel.name" @mouseup.left="goToChannel(channel)"
                            @mousedown.middle="goToChannel(channel, true)">
                            <td>
                                <pin-button class="float-end" :to="{ name: getRouteName('websocketChannel').value, params: { service: props.service.name, channel: channel.name } }" :label="channel.name" />
                                <router-link @click.stop class="row-link"
                                    :to="{ name: getRouteName('websocketChannel').value, params: { service: props.service.name, channel: channel.name } }">
                                    <template v-if="channel.parameters?.length">
                                        <span class="visually-hidden">Template </span>
                                        <span class="channel-name">
                                            <template v-for="(part, i) in splitTemplate(channel.name)" :key="i">
                                                <span v-if="part.param" class="param">{{ part.text }}</span>
                                                <template v-else>{{ part.text }}</template>
                                            </template>
                                        </span>
                                        <span class="visually-hidden">, </span>
                                        <span class="badge ms-2"
                                        :class="instanceCount(channel) ? 'bg-secondary' : 'bg-light text-muted'">
                                            {{ instanceLabel(channel) }}
                                        </span>
                                    </template>
                                    <span v-else>{{ channel.name }}</span>
                                </router-link>
                            </td>
                            <td>
                                <div>{{ channel.summary }}</div>
                            </td>
                            <td class="text-center">{{ lastMessage(channel) }}</td>
                            <td class="text-center">{{ channel.metrics.messages_total }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>