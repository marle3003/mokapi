<script setup lang="ts">
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRoute } from '@/router';
import type { ServiceResult } from '@/types/dashboard';
import { computed, ref, watch } from 'vue';
import Markdown from 'vue3-markdown-it'

const route = useRoute();
const { dashboard } = useDashboard()
const serviceName = route.params.service?.toString();
const serverName = route.params.server?.toString();
let data = ref<ServiceResult | null>(null);

const service = computed(() => {
    if (!data.value) {
        return undefined;
    }
    return data.value.service as KafkaService
})

const server = computed(() => {
    if (!service.value) {
        return undefined;
    }
    for (const server of service.value.servers) {
        if (server.name === serverName) {
            return server;
        }
    }
    return undefined;
})

watch(() => dashboard.value,
    (db, _, onCleanup) => {
        if (!serviceName) {
            return;
        }
        const res = db.getService(serviceName, 'kafka');
        data.value = res;

        onCleanup(() => res.close());
    },
    { immediate: true }
);
</script>

<template>
    <div v-if="server">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row mb-2">
                        <div class="col header">
                            <p id="name" class="label">Server Name</p>
                            <p aria-labelledby="name">{{ server.name }}</p>
                        </div>
                        <div class="col-2">
                            <p id="protocol" class="label">Protocol</p>
                            <p aria-labelledby="protocol">{{ server.protocol || '-' }}</p>
                        </div>
                        <div class="col">
                            <p id="cluster" class="label">Cluster</p>
                            <p>
                                <router-link :to="{
                                    name: getRouteName('kafkaService').value,
                                    params: { service: serviceName },
                                }" aria-labelledby="cluster">
                                    {{ serviceName }}
                                </router-link>
                            </p>
                        </div>
                    </div>
                    <div class="row mb-2" v-if="server.title">
                        <div class="col">
                            <p id="title" class="label">Title</p>
                            <p aria-labelledby="title">{{ server.title }}</p>
                        </div>
                    </div>
                    <div class="row mb-2" v-if="server.summary">
                        <div class="col">
                            <p id="summary" class="label">Summary</p>
                            <p aria-labelledby="summary">{{ server.summary }}</p>
                        </div>
                    </div>
                    <div class="row mb-2" v-if="server.description">
                        <div class="col">
                            <p id="description" class="label">Description</p>
                            <markdown :source="server.description" aria-labelledby="description" :html="true"></markdown>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="card-group" v-if="server.tags && server.tags.length > 0">
            <section class="card" aria-labelledby="tags">
                <div class="card-body">
                    <h2 id="tags" class="card-title text-center">Tags</h2>

                    <div class="table-responsive-sm mt-4">
                        <table class="table dataTable" aria-labelledby="tags">
                            <thead>
                                <tr>
                                    <th scope="col" class="text-left col-4">Name</th>
                                    <th scope="col" class="text-left col">Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="tag in server.tags" :key="tag.name">
                                    <td>{{ tag.name }}</td>
                                    <td>
                                        <markdown :source="tag.description" :html="true"></markdown>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </section>
        </div>
    </div>
    <loading v-if="!data || data.isLoading"></loading>
    <div v-if="data && !server && !data.isLoading">
        <message message="Kafka server not found"></message>
    </div>
</template>