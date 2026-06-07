<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';
import { useMqtt } from '@/composables/mqtt';

const props = defineProps<{
    service: MqttService,
}>()

const router = useRouter()
const { formatAddress, fromatVersion } = useMqtt();

const clients = computed(() => {
    if (!props.service || !props.service.clients) {
        return []
    }

    return props.service.clients.sort((c1: MqttClient, c2: MqttClient) => {
        return c1.clientId.localeCompare(c2.clientId)
    })
})

onMounted(() => {
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

function goToClient(client: MqttClient, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('mqttClient').value,
        params: {
            service: props.service.name,
            clientId: client.clientId,
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
        <table class="table dataTable selectable" aria-label="Clients">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-6">Client Id</th>
                    <th scope="col" class="text-left col-3">Address</th>
                    <th scope="col" class="text-left col-3">Protocol Version</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="c in clients" :key="c.clientId" @click.left="goToClient(c)"
                    @mousedown.middle="goToClient(c, true)">
                    <td>
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('mqttClient').value, params: { service: service.name, clientId: c.clientId } }">
                            {{ c.clientId }}
                        </router-link>
                    </td>
                    <td>{{ formatAddress(c.address) }}</td>
                    <td>{{ fromatVersion(c.protocolVersion) }}</td>
                </tr>
            </tbody>
        </table>
    </div>
</template>