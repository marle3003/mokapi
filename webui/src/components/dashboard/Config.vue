<script setup lang="ts">
import { computed, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import ConfigCard from '@/components/dashboard/ConfigCard.vue'
import SourceView from './SourceView.vue'
import { useDashboard } from '@/composables/dashboard'

const configId = useRoute().params.id as string
const { dashboard } = useDashboard()
const { config, isLoading, close } = dashboard.value.getConfig(configId);
const { data, isLoading: isDataLoading, filename, close: closeData } = dashboard.value.getConfigData(configId);

const { format } = usePrettyDates()
const { getContentType, formatLanguage } = usePrettyLanguage()
const defaultContentType = 'plain/text'

function isInitLoading() {
    return isLoading.value && !config.value && isDataLoading.value
}

const contentType = computed(() => {
    const name = filename.value
    if (!name) {
        return defaultContentType
    }
    const ct = getContentType(name)
    if (ct) {
        return ct
    }
    return defaultContentType
})

function toString(arg: any): string {
    if (typeof arg === 'string') {
        return arg
    }
    return JSON.stringify(arg)
}

onUnmounted(() => {
    close()
    closeData()
})
const provider = computed(() => {
    if (!config.value) {
        return '';
    }
    switch (config.value.provider.toLocaleLowerCase()) {
        case 'file': return 'File';
        case 'http': return 'HTTP';
        case 'git': return 'GIT';
        case 'npm': return 'NPM';
    }
    return '';
})
</script>

<template>
    <div v-if="config">
        <div class="card-group">
            <section class="card" aria-label="Info">
                <div class="card-body">
                    <div class="row">
                        <div class="col header mb-3">
                            <p id="url" class="label">URL</p>
                            <p aria-labelledby="url">{{ config.url }}</p>
                        </div>
                        <div class="col text-end">
                            <span v-if="config.tags" v-for="tag of config.tags" class="badge bg-primary ms-1">
                                {{ tag }}
                            </span>
                            <span class="badge bg-secondary ms-1">Config</span>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col">
                            <p id="provider" class="label">Provider</p>
                            <p aria-labelledby="provider">
                                <a :href="'/docs/configuration/dynamic/' + config.provider">{{ provider }}</a>
                            </p>
                        </div>
                        <div class="col">
                            <p id="lastModified" class="label">Last Modified</p>
                            <p aria-labelledby="lastModified">{{ format(config.time) }}</p>
                        </div>
                    </div>
                </div>
            </section>
        </div>
        <section class="card-group" v-if="data" aria-label="Content">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <source-view
                            :source="{ preview: { content: formatLanguage(toString(data), contentType), contentType: contentType } }"
                            :url="dashboard.getConfigDataUrl(configId)" :filename="filename" />
                    </div>
                </div>
            </div>
        </section>
        <div class="card-group" v-if="config.refs">
            <config-card :configs="config.refs" title="References" />
        </div>
    </div>
    <loading v-if="isInitLoading()"></loading>
    <div v-if="!config && !isLoading">
        <message message="Config not found"></message>
    </div>
</template>