<script setup lang="ts">
import OptionFilterEditor from './OptionFilterEditor.vue'
import type { MultiFilterItem } from '@/composables/http-request-filter'

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'TRACE', 'QUERY', 'CUSTOM']
const item = defineModel<MultiFilterItem & { custom?: string }>({ required: true })
</script>

<template>
    <option-filter-editor v-model="item" label="Method" :options="methods">
        <input v-if="item.value.includes('CUSTOM')" type="text" class="form-control form-control-sm mt-2"
            v-model="item.custom" placeholder="LINK CONNECT..." aria-label="Custom method">
    </option-filter-editor>
</template>
