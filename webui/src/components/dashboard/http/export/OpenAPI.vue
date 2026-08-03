<script setup lang="ts">
import { ref, computed, onMounted, useTemplateRef, watch } from 'vue'
import { Modal } from 'bootstrap'
import { transformPath } from '@/composables/fetch'
import { useRoute } from '@/router'

const props = defineProps({
    visible: { type: Boolean, default: false },
    serviceName: { type: String, required: true }
})
const emit = defineEmits(['update:visible'])

const route = useRoute()
const isLiveMode = computed(() => route.meta.mode === 'live')
const modalEl = useTemplateRef('modalEl')
let modalInstance: Modal | null = null

const formatOptions = [{ value: 'json', label: 'JSON' }, { value: 'yaml', label: 'YAML' }]
const selectedFormat = ref('json')
const encodedName = computed(() => encodeURIComponent(props.serviceName))
const downloadUrl = computed(() => {
    const base = `/api/services/http/${encodedName.value}`
    return transformPath(`${base}/openapi.${selectedFormat.value}`)
})
const copied = ref(false)

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
                    <h5 class="modal-title" id="export-openapi-title">Export OpenAPI Specification</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div v-if="!isLiveMode" class="alert alert-primary py-2 px-3 mb-3" role="alert">
                        This is a demo dashboard — OpenAPI exports aren't available here.
                        <a href="/docs/get-started/installation" >Install Mokapi</a> to try this with your own API.
                    </div>
                    <div v-else>
                        <div class="row">
                            <div class="col-2">
                                <p class="label">Format</p>
                            </div>
                            <div class="col">
                                <div class="form-check form-check-inline" v-for="opt in formatOptions" :key="opt.value">
                                    <input
                                        class="form-check-input"
                                        type="radio"
                                        :id="`fmt-${opt.value}`"
                                        :value="opt.value"
                                        v-model="selectedFormat"
                                    >
                                    <label class="form-check-label" :for="`fmt-${opt.value}`">
                                        {{ opt.label }}
                                    </label>
                                </div>
                            </div>
                        </div>
                        <div class="row">
                            <label for="download-url" class="col-2 col-form-label col-form-label-sm">URL</label>
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
                        :class="{ disabled: !isLiveMode }"
                        :href="downloadUrl"
                        :download="isLiveMode ? props.serviceName + '.' + selectedFormat : null"
                        @click="isLiveMode ? close() : $event.preventDefault()"
                    >
                        <i class="bi bi-download me-1"></i>Download
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>