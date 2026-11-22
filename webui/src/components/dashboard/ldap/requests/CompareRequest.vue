<script setup lang="ts">
defineProps<{
    request: LdapCompareRequest
}>()
</script>

<template>
    <div class="row mb-2">
        <div class="col-2">
            <p class="label">Attribute</p>
            <p>{{ request.attribute }}</p>
        </div>
        <div class="col">
            <p class="label">Value</p>
            <p>{{ request.value }}</p>
        </div>
    </div>
</template>
