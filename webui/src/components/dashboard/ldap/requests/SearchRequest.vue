<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    request: LdapSearchRequest
}>()

const attributes = computed(() => props.request.attributes?.join(', ') ?? '')
</script>

<template>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Base DN</p>
            <p>{{ request.baseDN }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col-2">
            <p class="label">Scope</p>
            <p>{{ request.scope }}</p>
        </div>
        <div class="col-2">
            <p class="label">Size Limit</p>
            <p>{{ request.sizeLimit > 0 ? request.sizeLimit : 'no limit' }}</p>
        </div>
        <div class="col-2">
            <p class="label">Time Limit</p>
            <p>{{ request.timeLimit > 0 ? request.timeLimit + ' [s]' : 'no limit' }}</p>
        </div>
    </div>
    <div class="row mb-2" v-if="attributes.length > 0">
        <div class="col">
            <p class="label">Attributes</p>
            <p>{{ attributes }}</p>
        </div>
    </div>
</template>
