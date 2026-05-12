<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { computed } from 'vue'
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';

const props = defineProps<{
    service: KafkaService,
    topic?: KafkaTopic
}>()

const router = useRouter()
const { format } = usePrettyDates()

const groups = computed(() => {
    if (!props.topic) {
        return props.service.groups
    }
    return props.topic.groups
})
function lastRebalancing(group: KafkaGroupInfo) {
  const timestamp = group.metrics.kafka_rebalance_timestamp;
  if (!timestamp) {
    return '-'
  }
  return format(timestamp)
}

function goToGroup(group: KafkaGroupInfo, openInNewTab = false){
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
function lags(group: KafkaGroupInfo) {
    if (!props.topic) {
        return '-'
    }
    const topicMetrics = group.metrics.topics[props.topic.name]
    if (!topicMetrics) {
        return '-'
    }
    let result = 0;
    for (const partition of topicMetrics) {
        if (!partition.kafka_consumer_group_lag) {
            continue
        }
        result += partition.kafka_consumer_group_lag
    }
    return result
}
</script>

<template>
    <div class="table-responsive-sm">
        <table class="table dataTable selectable" :aria-label="props.topic ? 'Topic Groups' : 'Cluster Groups'">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-2">Name</th>
                    <th scope="col" class="text-left col-1">State</th>
                    <th scope="col" class="text-left col-2">Protocol</th>
                    <th scope="col" class="text-center col-2">Generation</th>
                    <th scope="col" class="text-center col-2">Last Rebalancing</th>
                    <th scope="col" class="text-center col-2">Members</th>
                    <th scope="col" class="text-center col-1" v-if="topic">Lag</th>
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
                    <td class="text-center">{{ group.members }}</td>
                    <td v-if="topic" class="text-center">
                        {{ lags(group) }}
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>