<script setup lang="ts">
import { Modal } from 'bootstrap'
import { onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'

const show = defineModel<boolean>('show', { required: true })
const image = defineModel<HTMLImageElement>('image')

const modalEl = useTemplateRef('modalEl')
const dialog = ref<Modal>()

onMounted(() => {
  if (!modalEl.value) return

  dialog.value = new Modal('#imageDialog', {})

  modalEl.value.addEventListener('hidden.bs.modal', () => {
    show.value = false
  })
})

onBeforeUnmount(() => {
  dialog.value?.dispose()
})

watch(show, (value) => {
  if (!dialog.value) return

  value ? dialog.value.show() : dialog.value.hide()
})
</script>

<template>
  <div class="modal fade" id="imageDialog" tabindex="-1" aria-hidden="true" ref="modalEl">
    <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-body" v-if="image">
          <img class="img-fluid shadow rounded" :src="image.src" style="width:100%" />
          <div class="pt-2" style="text-align:center; font-size:0.9rem;">
            {{ image.alt }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>