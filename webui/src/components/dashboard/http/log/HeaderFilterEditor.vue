<script setup lang="ts">
import { computed, ref } from 'vue'
import RegexInput from '@/components/RegexInput.vue'
import type { FilterItem, HeaderFilter } from '@/composables/http-request-filter'

defineProps<{ label: string }>()
const headers = defineModel<FilterItem<HeaderFilter[]>>({ required: true })

// eslint-disable-next-line @typescript-eslint/no-explicit-any -- template function refs hand over an element or a component instance
const valueInputs = ref<any[]>([])
const regexErrors = computed(() => valueInputs.value.filter(input => input?.regexError).map(input => input.regexError))

function onHeaderInput(index: number) {
    const rows = headers.value.value
    const last = rows[rows.length - 1]

    // The last row is always an empty one for the next header.
    if (index === rows.length - 1 && (last.name?.trim() !== '' || last.value?.trim())) {
        rows.push({ name: '', value: '' })
    }

    for (let i = 0; i < rows.length - 1; i++) {
        if (!rows[i].name?.trim() && !rows[i].value?.trim()) {
            rows.splice(i, 1)
            break
        }
    }
}

function removeRow(index: number) {
    const rows = headers.value.value
    rows.splice(index, 1)

    const last = rows[rows.length - 1]
    if (last && last.name) {
        rows.push({ name: '', value: '' })
    }
}
</script>

<template>
    <div v-for="(row, i) in headers.value" :key="i" class="row g-1 mb-2 align-items-center">
        <div class="col">
            <input type="text" class="form-control form-control-sm" v-model="row.name" placeholder="Name"
                @input="onHeaderInput(i)" :aria-label="`${label} Name ${i + 1}`">
        </div>
        <div class="col">
            <RegexInput v-model="row.value" :ref="el => valueInputs[i] = el" placeholder="Value [Regex]"
                @input="onHeaderInput(i)" :aria-label="`${label} Value ${i + 1}`" />
        </div>
        <div class="col-auto" style="width: 2rem;">
            <button v-if="i < headers.value.length - 1" type="button" class="btn btn-outline-danger btn-sm"
                @click="removeRow(i)" :aria-label="`Remove ${label} ${i + 1}`">
                <i class="bi bi-x"></i>
            </button>
        </div>
    </div>
    <div class="invalid-feedback" style="display: inline;" v-if="regexErrors.length > 0">
        Invalid regular expression:
        <ul>
            <li v-for="(err, errIndex) in regexErrors" :key="errIndex">{{ err }}</li>
        </ul>
    </div>
</template>
