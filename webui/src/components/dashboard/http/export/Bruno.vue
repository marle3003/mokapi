<script setup lang="ts">
import { ref, computed, onMounted, useTemplateRef, watch } from 'vue'
import { useRoute } from '@/router'
import { Modal } from 'bootstrap'
import { useDashboard } from '@/composables/dashboard'

const props = defineProps({
    visible: { type: Boolean, default: false },
    serviceName: { type: String, required: true }
})
const emit = defineEmits(['update:visible'])
const route = useRoute()
const isLiveMode = computed(() => route.meta.mode === 'live')

const modalEl = useTemplateRef('modalEl')
let modalInstance: Modal | null = null

const { dashboard } = useDashboard()
const downloadUrl = computed(() => {
    if (!dashboard.value) {
        return ''
    }
    const params = new URLSearchParams({
        host: baseUrl.value,
        folderArrangement: folderArrangement.value,
        itemName: itemName.value
    })
    const url = dashboard.value.getBrunoCollectionUrl(props.serviceName, params)
    return url.value || ''
})
const copied = ref(false)
const baseUrl = ref(window.location.host)
const folderArrangement = ref('tags')
const itemName = ref('summary')

onMounted(() => {
    if (!modalEl.value) {
        console.error('Modal element not found')
        return
    }
    modalInstance = new Modal(modalEl.value)
    modalEl.value.addEventListener('hidden.bs.modal', () => emit('update:visible', false))
})

watch(() => props.visible, v => {
    v ? modalInstance?.show() : modalInstance?.hide()
})

function close() {
    emit('update:visible', false)
}
function copyToClipboard(event: MouseEvent) {
    event.preventDefault()
    
    navigator.clipboard.writeText(downloadUrl.value)
        .then(() => {
            copied.value = true
            
            setTimeout(() => {
                copied.value = false
            }, 2000)
        })
        .catch(err => {
            console.error('Fehler beim Kopieren: ', err)
        })
}
</script>

<template>
    <div class="modal fade" id="export-bruno" tabindex="-1" ref="modalEl" aria-hidden="true" aria-labelledby="export-bruno-title">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="export-bruno-title">Export Bruno Collection</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div v-if="!isLiveMode" class="alert alert-primary py-2 px-3 mb-3 small" role="alert">
                        This is a demo dashboard — OpenAPI exports aren't available here.
                        <a href="/docs/get-started/installation" >Install Mokapi</a> to try this with your own API.
                    </div>
                    <p class="text-muted small mb-3">
                        <a href="https://www.usebruno.com" target="_blank" rel="noopener">Bruno</a>
                        is a free, open-source, Git-friendly API client — think Postman, but your
                        collections are just files you can version alongside your code.
                    </p>
                    <div v-if="isLiveMode">
                        <div class="row mb-3">
                            <label for="bruno-base-url" class="col-3 col-form-label col-form-label-sm">
                                Base URL
                                <i
                                    class="bi bi-question-circle text-muted ms-1"
                                    data-bs-toggle="tooltip"
                                    data-bs-placement="top"
                                    title="Used to complete server URLs in the spec that are relative — missing a scheme, host, or port (e.g. /api/books). Has no effect on server URLs that already define a host."
                                    style="cursor: help;"
                                ></i>
                            </label>
                            <div class="col">
                                <input
                                    type="text"
                                    id="bruno-base-url"
                                    class="form-control form-control-sm"
                                    v-model="baseUrl"
                                >
                            </div>
                        </div>

                        <div class="row mb-3">
                            <label for="bruno-folder-arrangment" class="col-3 col-form-label col-form-label-sm">
                                Folder Arrangement
                                <i
                                    class="bi bi-question-circle text-muted ms-1"
                                    data-bs-toggle="tooltip"
                                    data-bs-placement="top"
                                    title="Choose whether folders should be created based on the paths or tags of the specification."
                                    style="cursor: help;"
                                ></i>
                            </label>
                            <div class="col">
                                <select class="form-select form-select-sm" v-model="folderArrangement" id="bruno-folder-arrangment">
                                    <option value="tags" selected>Tags</option>
                                    <option value="paths">Paths</option>
                                </select>
                            </div>
                        </div>

                        <div class="row mb-3">
                            <label for="bruno-item-name" class="col-3 col-form-label col-form-label-sm">
                                Item Name
                                <i
                                    class="bi bi-question-circle text-muted ms-1"
                                    data-bs-toggle="tooltip"
                                    data-bs-placement="top"
                                    title="Preferred source for each request's name. Falls back to operationId, then path, if the chosen field is empty."
                                    style="cursor: help;"
                                ></i>
                            </label>
                            <div class="col">
                                <select class="form-select form-select-sm" v-model="itemName" id="bruno-item-name">
                                    <option value="summary" selected>Summary</option>
                                    <option value="path">Path</option>
                                </select>
                            </div>
                        </div>

                        <div class="row">
                            <label for="bruno-download-url" class="col-3 col-form-label col-form-label-sm">Download URL</label>
                            <div class="col">
                                <div class="input-group mb-3">
                                    <input type="url" class="form-control form-control-sm" id="bruno-download-url" :value="downloadUrl" readonly>
                                    <button class="btn btn-sm btn-outline-secondary" type="button" @click="copyToClipboard"
                                        aria-label="Copy download URL to clipboard"
                                    >
                                        <i class="bi bi-copy" aria-hidden="true"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                </div>
                <div class="modal-footer d-flex align-items-center justify-content-between p-1">
                    <div class="text-success small" 
                        :class="{ 'opacity-0': !copied, 'opacity-100': copied }" 
                        style="transition: opacity 0.2s ease-in-out;">
                        ✓ Copied to clipboard!
                    </div>

                    <a
                        class="btn btn-primary btn-sm"
                        :href="downloadUrl"
                        :download="props.serviceName + '.yaml'"
                        @click="close"
                    >
                        <i class="bi bi-download me-1"></i>Download
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>