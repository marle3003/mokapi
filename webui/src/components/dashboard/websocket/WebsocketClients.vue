<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';
import { useWebsocket } from '@/composables/websocket';

const props = defineProps<{
    service: WebsocketService,
}>()

const router = useRouter()
const { formatAddress } = useWebsocket();

const clients = computed(() => {
    if (!props.service || !props.service.clients) {
        return []
    }

    return props.service.clients.sort((c1: WebsocketClient, c2: WebsocketClient) => {
        return c1.id.localeCompare(c2.id)
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

function goToClient(client: WebsocketClient, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('websocketClient').value,
        params: {
            service: props.service.name,
            clientId: client.id,
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
                    <th scope="col" class="text-left col-3">Address</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="c in clients" :key="c.id" @click.left="goToClient(c)"
                    @mousedown.middle="goToClient(c, true)">
                    <td>
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('websocketClient').value, params: { service: service.name, clientId: c.id } }">
                            {{ formatAddress(c.address) }}
                        </router-link>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>