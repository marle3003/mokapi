<script setup lang="ts">
import { ref, useId, watch } from 'vue';


const model = defineModel<string | undefined>();
const id = useId();
const regex = ref<RegExp | undefined>()
const regexError = ref<unknown>(null)
defineExpose({ regex, regexError })

let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(model, () => {
    if (debounceTimer) {
        clearTimeout(debounceTimer)
    }

    try {
        const v = model.value?.trim();
        if (!v) {
            model.value = undefined
            regexError.value = null
            return
        }
        regex.value = new RegExp(v)
        regexError.value = null
    } catch (error) {
        debounceTimer = setTimeout(() => {
            regexError.value = error;
            regex.value = undefined;
        }, 1500)
    }
    
}, { immediate: true })
</script>

<template>
  <input  type="text" class="form-control form-control-sm" :id="id" v-model="model" :class="{ 'is-invalid': regexError }">
</template>