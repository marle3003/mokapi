<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue'
import { Modal } from 'bootstrap'

const props = defineProps<{
    response: LdapSearchResponse
}>()

// A search that is still running has no results yet. The list is sorted as a copy, the event stays as it is.
const results = computed(() => [...(props.response.results ?? [])].sort(
    (r1, r2) => r1.dn.toLowerCase().localeCompare(r2.dn.toLowerCase())
))

const selected = ref<{ dn: string, keys: string[], attributes: { [name: string]: string[] } }>()
const dialogRef = useTemplateRef('dialogRef')
let dialog: Modal | undefined

function showResult(result: LdapSearchResult) {
    if (getSelection()?.toString()) {
        return
    }
    selected.value = {
        dn: result.dn,
        keys: Object.keys(result.attributes).sort((x, y) => x.localeCompare(y)),
        attributes: result.attributes
    }
    dialog?.show()
}

onMounted(() => {
    if (dialogRef.value) {
        dialog = new Modal(dialogRef.value)
    }
})
onBeforeUnmount(() => dialog?.dispose())
</script>

<template>
    <table class="table dataTable selectable" aria-label="Results">
        <thead>
            <tr>
                <th scope="col" class="text-left col">DN</th>
                <th scope="col" class="text-center col-2">Attributes</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="item of results" :key="item.dn" @click="showResult(item)">
                <td>{{ item.dn }}</td>
                <td class="text-center">{{ Object.keys(item.attributes).length }}</td>
            </tr>
        </tbody>
    </table>

    <!-- The panel is sticky and so a stacking context of its own: inside it the dialog would sit below the backdrop. -->
    <teleport to="body">
        <!-- no fade: Bootstrap moves the focus into the dialog only when the animation has ended, until then the arrow keys would act on the log -->
        <div class="modal" tabindex="-1" aria-hidden="true" aria-labelledby="modal-response-title" ref="dialogRef">
            <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
                <div class="modal-content" v-if="selected">
                    <div class="modal-header">
                        <h6 id="modal-response-title" class="modal-title">DN: {{ selected.dn }}</h6>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <div class="row p-2">
                            <table class="table dataTable" aria-labelledby="modal-response-title">
                                <thead>
                                    <tr>
                                        <th scope="col" class="text-left col-4">Type</th>
                                        <th scope="col" class="text-left col">Value</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="name in selected.keys" :key="name">
                                        <td>{{ name }}</td>
                                        <td v-html="selected.attributes[name]?.join('<br />')"></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </teleport>
</template>
