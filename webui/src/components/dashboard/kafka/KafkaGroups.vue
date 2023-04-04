<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { type PropType, onMounted } from 'vue';
import { Popover } from 'bootstrap'
import { useMetrics } from '@/composables/metrics';

const props = defineProps({
    service: { type: Object as PropType<KafkaService>, required: true },
    topicName: { type: String, required: false }
})

const format = usePrettyDates().format
const {sum} = useMetrics()

function memberInfo(member: KafkaMember): string {
  return `<p class="label">Address</p>
           <p>${member.addr}</p>
           <p class="label">Client Software</p>
           <p>${member.clientSoftwareName} ${member.clientSoftwareVersion}</p>
           <p class="label">Last Heartbeat</p>
           <p>${format(member.heartbeat)}</p>`
}

function getGroups(): KafkaGroup[] {
    if (!props.topicName) {
        return props.service.groups
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
  new Popover(document.body, {
      selector: ".member[data-bs-toggle='popover']",
      customClass: 'dashboard-popover',
      html: true,
      trigger: 'hover',
      content: function(this: HTMLElement): string {
        return this.nextElementSibling?.outerHTML ?? ''
      }
    })
})
</script>

<template>
    <table class="table dataTable selectable">
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
                    <div v-for="member in group.members">
                        <div class="member" data-bs-toggle="popover" data-bs-placement="right" >{{ member.name }} <i class="bi bi-info-circle"></i></div>
                        <span style="display:none" v-html="memberInfo(member)"></span>
                    </div>
                </td>
                <td v-if="topicName" class="text-center">
                    {{ sum(service.metrics, 'kafka_consumer_group_lag', {name: 'topic', value: topicName}, {name: 'group', value: group.name }) }}
                </td>
            </tr>
        </tbody>
    </table>
</template>