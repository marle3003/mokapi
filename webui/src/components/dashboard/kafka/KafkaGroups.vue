<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useMetrics } from '@/composables/metrics'

const props = defineProps<{
    service: KafkaService,
    topicName?: string
}>()

const { format } = usePrettyDates()
const { sum } = useMetrics()

function memberInfo(member: KafkaMember): string {
    return `<div aria-label="${member.name}">
            <p id="${member.name}-address" class="label">Address</p>
            <p aria-labelledby="${member.name}-address">${member.addr}</p>
            <p id="${member.name}-client-software" class="label">Client Software</p>
            <p aria-labelledby="${member.name}-client-software">${member.clientSoftwareName} ${member.clientSoftwareVersion}</p>
            <p id="${member.name}-last-heartbeat" class="label">Last Heartbeat</p>
            <p aria-labelledby="${member.name}-last-heartbeat">${format(member.heartbeat)}</p>
            <p id="${member.name}-partitions" class="label">Partitions</p>
            <p aria-labelledby="${member.name}-partitions">${member.partitions.join(', ')}</p>
            </div>`
}

function getGroups(): KafkaGroup[] {
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
}
onMounted(()=> {
    const elements = document.querySelectorAll('.has-popover')
    const popovers = [...elements].map(x => {
        new Popover(x, {
            customClass: 'custom-popover',
            trigger: 'hover',
            html: true,
            placement: 'left',
            content: () => x.querySelector('span')?.innerHTML ?? '',
        })
    })
})
</script>

<template>
    <table class="table dataTable">
        <caption class="visually-hidden">{{ props.topicName ? 'Topic Groups' : 'Cluster Groups' }}</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">State</th>
                <th scope="col" class="text-left">Protocol</th>
                <th scope="col" class="text-left">Coordinator</th>
                <th scope="col" class="text-left">Leader</th>
                <th scope="col" class="text-left">Members</th>
                <th scope="col" class="text-center" v-if="topicName">Lag</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="group in getGroups()" :key="group.name">
                <td>{{ group.name }}</td>
                <td>{{ group.state }}</td>
                <td>{{ group.protocol }}</td>
                <td>{{ group.coordinator }}</td>
                <td>{{ group.leader }}</td>
                <td>
                    <ul class="members">
                        <li v-for="member in group.members" class="has-popover">
                            {{ member.name }} <i class="bi bi-info-circle"></i>
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
</template>

<style scoped>
ul.members {
    list-style: none; 
    padding: 0;
    margin: 0;
}
ul.members li {
    padding-right: 0.5em;
}
</style>