<script setup lang="ts">
import { useConfig } from '@/composables/configs'
import { computed, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import ConfigCard from '@/components/dashboard/ConfigCard.vue'
import SourceView from './SourceView.vue'

const configId = useRoute().params.id as string
const { fetch, fetchData, getDataUrl } = useConfig()
const { format } = usePrettyDates()
const { getContentType } = usePrettyLanguage()
const defaultContentType = 'plain/text'

const { config, isLoading, close } = fetch(configId)
const { data, isLoading: isLoadingData, close: closeData } = fetchData(configId)

function isInitLoading() {
    return isLoading.value && !config.value && isLoadingData
}

const contentType = computed(() => {
    if (!config.value) {
        return defaultContentType
    }
    const ct = getContentType(config.value?.url)
    if (ct) {
        return ct
    }
    return defaultContentType
})

const filename = computed(() => {
    if (!config.value) {
        return ''
    }
    const url = config.value.url
    return url.substring(url.lastIndexOf('/')+1);
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
</script>

<template>
  <div v-if="config">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col-11 header mb-3">
                            <p id="url" class="label">URL</p>
                            <p aria-labelledby="url">{{ config.url }}</p>
                        </div>
                        <div class="col-1 text-end">
                            <span class="badge bg-secondary">Config</span>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col">
                            <p id="provider" class="label">Provider</p>
                            <p aria-labelledby="provider">
                                <a href="/docs/configuration/providers/file">{{ config.provider }}</a>
                            </p>
                        </div>
                        <div class="col">
                            <p id="lastModified" class="label">Last Modified</p>
                            <p aria-labelledby="lastModified">{{ format(config.time) }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <section class="card-group" v-if="data" aria-label="Content">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <source-view :source="toString(data)" :content-type="contentType" :url="getDataUrl(configId)" :filename="filename" height="500px" />
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
        <message message="HTTP Request not found"></message>
    </div>
</template>