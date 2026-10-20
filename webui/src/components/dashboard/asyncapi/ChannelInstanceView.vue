<script setup lang="ts" generic="Channel extends BaseChannel">
import { computed, reactive } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import router from '@/router'
import { getRouteName, } from '@/composables/dashboard'
import { usePageSize } from '@/composables/paging'
import Paging from '../Paging.vue'

const props = defineProps<{
    serviceName: string
    channel: Channel
    instances: ChannelInstance[]
    instancesLabel: string
    instanceLabel: string
    channelRoute: string
}>()

const { format } = usePrettyDates()

const paging = reactive({
    current: 1,
    size: usePageSize(25),
})

const data = computed(() => {
    return props.instances.slice((paging.current - 1) * paging.size, paging.current * paging.size)
})

function goToChannel(instance: string, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to: { name: string, params: Record<string, string> } = {
        name: getRouteName(props.channelRoute).value,
        params: { service: props.serviceName }
    }
    if (props.channelRoute === 'websocketChannel') {
        to.params.channel = instance
    } else {
        to.params.topic = instance
    }

    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}

function lastMessage(channel: ChannelInstance) {
    if (!channel.metrics?.message_timestamp) {
        return '-'
    }
    return format(channel.metrics.message_timestamp)
}
</script>

<template>
    <section class="card" aria-labelledby="channels">
        <div class="card-body">
            <h2 id="channels" class="card-title text-center">{{ instancesLabel }}</h2>

            <div class="table-responsive">
                <table class="table dataTable selectable" aria-labelledby="channels">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-3">{{ instanceLabel }}</th>
                            <th scope="col" class="text-center col-2">Last Message</th>
                            <th scope="col" class="text-center col-1">Messages</th>

                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="instance in data" :key="channel.name" @mouseup.left="goToChannel(instance.name)"
                            @mousedown.middle="goToChannel(instance.name, true)">
                            <td>
                                <router-link @click.stop class="row-link"
                                    :to="{ name: getRouteName(channelRoute).value, params: { service: props.serviceName, ...(channelRoute === 'websocketChannel' ? {channel: instance.name} : { topic: instance.name }) } }">
                                    {{ instance.name }}
                                </router-link>
                            </td>
                            <td class="text-center">{{ lastMessage(instance) }}</td>
                            <td class="text-center">{{ instance.metrics?.messages_total ?? 0 }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
            
            <Paging v-if="instances"
                :count="data.length"
                :total="instances.length"
                v-model:current="paging.current"
                v-model:size="paging.size"
            />
        </div>
    </section>
</template>