<script setup lang="ts">
import { useRouter } from 'vue-router';
import { type PropType, computed, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { getRouteName, useDashboard } from '@/composables/dashboard';

const props = defineProps({
    service: { type: Object as PropType<MailService> },
})
let data: SmtpEventData | null

const labels: Label[] = []
if (props.service) {
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const { dashboard } = useDashboard()
const { events, close } = dashboard.value.getEvents('mail', ...labels)
const { format, duration } = usePrettyDates()

function goToMail(evt: ServiceEvent, openInNewTab = false){
    const data = eventData(evt)

    const to = {
        name: getRouteName('smtpMail').value,
        params: {id: data.messageId},
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
function eventData(event: ServiceEvent): SmtpEventData{
    return <SmtpEventData>event.data
}

const filteredEvents = computed(() => {
    if (!events.value) {
        return []
    }
    return events.value.filter(e => {
        const data = eventData(e)
        return data.messageId
    })
})

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <h2 class="card-title text-center">Recent Mails</h2>
            <table class="table dataTable selectable">
                <caption class="visually-hidden">Recent Mails</caption>
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 35%">Subject</th>
                        <th scope="col" class="text-left" style="width: 20%">From</th>
                        <th scope="col" class="text-left" style="width: 20%">To</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center" style="width:10%">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in filteredEvents" :key="event.id" :set="data = eventData(event)" @mouseup.left="goToMail(event)" @mousedown.middle="goToMail(event, true)">
                        <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('smtpMail').value, params: {id: data.messageId} }">
                                {{ data.subject }}
                            </router-link>
                        </td>
                        <td>{{ data.from }}</td>
                        <td class="address-list">
                            <div v-for="address in data.to">
                            {{ address }}
                            </div>
                        </td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(data.duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style scoped>
.address-list ul {
    margin-bottom: 0;
}
</style>