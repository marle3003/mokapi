<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { computed, onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useMetrics } from '@/composables/metrics'
import { useKafka } from '@/composables/kafka';
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';

const props = defineProps<{
    service: KafkaService,
    topicName?: string
}>()

const router = useRouter()
const { format } = usePrettyDates()
const { sum } = useMetrics()
const { clientSoftware } = useKafka();

function memberInfo(member: KafkaMember): string {
    let addition = ''
    if (props.topicName) {
        addition = `<p id="${member.name}-partitions" class="label">Partitions</p><p aria-labelledby="${member.name}-partitions">${member.partitions[props.topicName]?.join(', ')}</p>`
    } else {
        const topics = Object.keys(member.partitions).sort((x, y) => x.localeCompare(y)).join(',<br />')
        addition = `<p id="${member.name}-topics" class="label">Topics</p><p aria-labelledby="${member.name}-topics">${topics}</p>`
    }
    return `<div aria-label="${member.name}">
            <p id="${member.name}-address" class="label">Address</p>
            <p aria-labelledby="${member.name}-address">${member.addr}</p>
            <p id="${member.name}-client-software" class="label">Client Software</p>
            <p aria-labelledby="${member.name}-client-software">${clientSoftware(member)}</p>
            <p id="${member.name}-last-heartbeat" class="label">Last Heartbeat</p>
            <p aria-labelledby="${member.name}-last-heartbeat">${format(member.heartbeat)}</p>
            ${addition}
            </div>`
}

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

onMounted(()=> {
    const elements = document.querySelectorAll('.has-popover')
    const popovers = [...elements].map(x => {
        new Popover(x, {
            customClass: 'custom-popover',
            trigger: 'hover',
            html: true,
            placement: 'left',
            content: () => x.querySelector('span:not(.bi)')?.innerHTML ?? '',
        })
    })
})

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
                    <th scope="col" class="text-left col-2">Coordinator</th>
                    <th scope="col" class="text-left col-2">Leader</th>
                    <th scope="col" class="text-left col-2">Members</th>
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
                    <td v-html="group.coordinator.replace(/([^:]*):(.*)/g, '$1<wbr>:$2')"></td>
                    <td>{{ group.leader }}</td>
                    <td>
                        <ul class="members">
                            <li v-for="member in group.members" class="has-popover">
                                {{ member.name }} <span class="bi bi-info-circle"></span>
                                <span style="display:none" v-html="memberInfo(member)"></span>
                            </li>
                            
                        </ul>
                    </td>
                    <td v-if="topicName" class="text-center">
                        {{ sum(service.metrics, 'kafka_consumer_group_lag', { name: 'topic', value: topicName }, { name: 'group', value: group.name }) }}
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>