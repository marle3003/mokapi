<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { computed } from 'vue'
import { useMetrics } from '@/composables/metrics'
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';

const props = defineProps<{
    service: KafkaService,
    topicName?: string
}>()

const router = useRouter()
const { format } = usePrettyDates()
const { sum, value } = useMetrics();

const groups = computed(() => {
    if (!props.topicName) {
        return props.service.groups
    }
    if (!props.service.groups) {
        return []
    }

    let result = []
    for (let group of props.service.groups) {
        if (group.topics?.includes(props.topicName)) {
            result.push(group)
        }
    }
    return result
})
function lastRebalancing(group: KafkaGroup) {
  const timestamp = value(props.service.metrics, 'kafka_rebalance_timestamp', { name: 'group', value: group.name });
  if (!timestamp) {
    return '-'
  }
  return format(timestamp)
}

function goToGroup(group: KafkaGroup, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaGroup').value,
        params: {
          service: props.service.name,
          group: group.name,
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
    <div class="table-responsive-sm">
        <table class="table dataTable selectable" :aria-label="props.topicName ? 'Topic Groups' : 'Cluster Groups'">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-2">Name</th>
                    <th scope="col" class="text-left col-1">State</th>
                    <th scope="col" class="text-left col-2">Protocol</th>
                    <th scope="col" class="text-center col-2">Generation</th>
                    <th scope="col" class="text-center col-2">Last Rebalancing</th>
                    <th scope="col" class="text-center col-2">Members</th>
                    <th scope="col" class="text-center col-1" v-if="topicName">Lag</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="group in groups" :key="group.name" @click.left="goToGroup(group)" @mousedown.middle="goToGroup(group, true)">
                    <td>
                        <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaGroup').value, params: { service: service.name, group: group.name }}">
                            {{ group.name }}
                        </router-link>
                    </td>
                    <td>{{ group.state }}</td>
                    <td v-html="group.protocol.replace(/([a-z])([A-Z])/g, '$1<wbr>$2')"></td>
                    <td class="text-center">{{ group.generation }}</td>
                    <td class="text-center">{{ lastRebalancing(group) }}</td>
                    <td class="text-center">{{ group.members.length }}</td>
                    <td v-if="topicName" class="text-center">
                        {{ sum(service.metrics, 'kafka_consumer_group_lag', { name: 'topic', value: topicName }, { name: 'group', value: group.name }) }}
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>