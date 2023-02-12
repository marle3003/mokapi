<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { type PropType, onMounted } from 'vue';
import { Popover } from 'bootstrap'

const format = usePrettyDates().format
const props = defineProps({
    service: { type: Object as PropType<KafkaService>, required: true },
    topicName: { type: String, required: false }
})
onMounted(()=> {
  new Popover(document.body, {
      selector: "[data-bs-toggle='popover']",
      html: true,
      customClass: 'dashboard-popover',
      trigger: 'hover'
    })
})
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
        if (group.topics.includes(props.topicName)) {
            result.push(group)
        }
    }
    return result
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Groups</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Name</th>
                        <th scope="col" class="text-left">State</th>
                        <th scope="col" class="text-left">Protocol</th>
                        <th scope="col" class="text-left">Coordinator</th>
                        <th scope="col" class="text-left">Leader</th>
                        <th scope="col" class="text-left">Members</th>
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
                          <div v-for="member in group.members" data-bs-toggle="popover" data-bs-placement="right" 
                            :title="member.name"
                            :data-bs-content="memberInfo(member)">{{ member.name }}</div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>