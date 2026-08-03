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
    const params: Record<string, string> = {}
    if (hostOverride.value) {
        params['host'] = hostOverride.value
    }
    const url = dashboard.value.getBrunoCollectionUrl(props.serviceName, params)
    return url.value || ''
})
const copied = ref(false)
const hostOverride = ref(window.location.host)

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
    <div class="modal fade" id="export-openapi" tabindex="-1" ref="modalEl" aria-hidden="true" aria-labelledby="export-openapi-title">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="export-openapi-title">Export Bruno Collection</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <p class="text-muted small mb-3">
                        <a href="https://www.usebruno.com" target="_blank" rel="noopener">Bruno</a>
                        is a free, open-source, Git-friendly API client — think Postman, but your
                        collections are just files you can version alongside your code.
                    </p>
                    <div v-if="isLiveMode">
                        <div class="row mb-3">
                            <label for="export-host" class="col-3 col-form-label col-form-label-sm">
                                Host
                                <i
                                    class="bi bi-question-circle text-muted ms-1"
                                    ref="tooltipEl"
                                    data-bs-toggle="tooltip"
                                    data-bs-placement="top"
                                    title="Fills in the host for server URLs without one (e.g. /api/books). Has no effect on server URLs that already specify a full host."
                                    style="cursor: help;"
                                ></i>
                            </label>
                            <div class="col">
                                <input
                                    type="text"
                                    id="export-host"
                                    class="form-control form-control-sm"
                                    v-model="hostOverride"
                                    placeholder="api.example.com"
                                >
                            </div>
                        </div>

                        <div class="row">
                            <label for="download-url" class="col-3 col-form-label col-form-label-sm">Download URL</label>
                            <div class="col">
                                <div class="input-group mb-3">
                                    <input type="url" class="form-control form-control-sm" id="download-url" :value="downloadUrl" readonly>
                                    <button class="btn btn-sm btn-outline-secondary" type="button" @click="copyToClipboard">
                                        <i class="bi bi-copy"></i>
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