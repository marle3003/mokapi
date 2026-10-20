<script setup lang="ts">
import { useRouter } from 'vue-router'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard';
import { computed } from 'vue';
import { useLocalStorage } from '@/composables/local-storage';
import { splitTemplate, instanceCount } from '@/composables/channel'

const props = defineProps<{
    service: WebsocketService,
}>()

const { format } = usePrettyDates()
const router = useRouter()
const tags = useLocalStorage<string[]>(`websocket-${props.service.name}-tags`, ['__all'])

const allTags = computed(() => {
    if (!props.service || !props.service.channels) {
        return [];
    }
    const result: Tag[] = [];
    const names: { [name: string]: any } = {}

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
const channels = computed(() => {
    if (!props.service.channels) {
        return []
    }
    let result = props.service.channels.sort(compareChannels)
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

function compareChannels(c1: WebsocketChannel, c2: WebsocketChannel) {
    const name1 = c1.name.toLowerCase()
    const name2 = c2.name.toLowerCase()
    return name1.localeCompare(name2)
}

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

            <fieldset class="text-center mt-3 mb-2" v-if="allTags.length > 1" aria-describedby="tag-help">

                <legend class="visually-hidden">
                    Filter channels by tags
                </legend>

                <p id="tag-help" class="visually-hidden">
                    Select one or more tags to filter the channels. Selecting “All” enables all tags.
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
                <table class="table dataTable selectable" aria-labelledby="channels">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-4">Path</th>
                            <th scope="col" class="text-left col-4">Summary</th>
                            <th scope="col" class="text-center col-2">Last Message</th>
                            <th scope="col" class="text-center col-1">Messages</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="channel in channels" :key="channel.name" @mouseup.left="goToChannel(channel)"
                            @mousedown.middle="goToChannel(channel, true)">
                            <td>
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