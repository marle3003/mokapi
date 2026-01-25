<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useKafka } from '@/composables/kafka';
import { useRouter } from '@/router';
import { getRouteName } from '@/composables/dashboard';

const props = defineProps<{
    service: KafkaService,
}>()

const router = useRouter()
const { clientSoftware, formatAddress } = useKafka();

const clients = computed(() => {
    if (!props.service || !props.service.clients) {
        return []
    }

    return props.service.clients.sort((c1: KafkaClient, c2: KafkaClient) => {
        return c1.clientId.localeCompare(c2.clientId)
    })
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

function goToClient(client: KafkaClient, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaClient').value,
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
    <section class="card" aria-labelledby="clients">
        <div class="card-body">
            <h2 id="clients" class="card-title text-center">Clients</h2>

                <div class="table-responsive-sm">
                    <table class="table dataTable selectable" aria-labelledby="clients">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-6">ClientId</th>
                                <th scope="col" class="text-left col-3">Address</th>
                                <th scope="col" class="text-left col-3">Software</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="c in clients" :key="c.clientId" @click.left="goToClient(c)" @mousedown.middle="goToClient(c, true)">
                                <td>
                                    <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaClient').value, params: { service: service.name, clientId: c.clientId }}">
                                        {{ c.clientId }}
                                    </router-link>
                                </td>
                                <td>{{ formatAddress(c.address) }}</td>
                                <td>{{ clientSoftware(c) }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
        </div>
    </section>
</template>