<script setup lang="ts">
import { computed } from 'vue'
import type { MultiFilterItem } from '@/composables/http-request-filter'

defineProps<{ label: string, options: string[] }>()
const item = defineModel<MultiFilterItem>({ required: true })

const isMulti = computed(() => item.value.state === 'Multi')

const single = computed({
    get: () => item.value.value[0] ?? '',
    set: (value: string) => { item.value.value = value ? [value] : [] },
})

function toggleMulti() {
    if (isMulti.value) {
        item.value.state = 'Single'
        item.value.value = item.value.value.slice(0, 1)
    } else {
        item.value.state = 'Multi'
    }
}
</script>

<template>
    <select v-if="!isMulti" class="form-select form-select-sm" v-model="single" :aria-label="label">
        <option v-for="option of options" :key="option">{{ option }}</option>
    </select>
    <div v-else role="group" :aria-label="label">
        <div class="form-check" v-for="option of options" :key="option">
            <input class="form-check-input" type="checkbox" :id="`${label}-${option}`" :value="option" v-model="item.value">
            <label class="form-check-label" :for="`${label}-${option}`">{{ option }}</label>
        </div>
    </div>
    <slot />
    <button type="button" class="btn btn-link btn-sm px-0 mt-1" @click="toggleMulti">
        {{ isMulti ? 'Select one' : 'Select several' }}
    </button>
</template>
