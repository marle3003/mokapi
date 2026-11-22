<script setup lang="ts">
defineProps<{
    request: LdapModifDNRequest
}>()
</script>

<template>
    <div class="row mb-2">
        <div class="col-2">
            <p class="label">New RDN</p>
            <p>{{ request.newRdn }}</p>
        </div>
        <div class="col-2">
            <p class="label">Delete Old RDN</p>
            <p>{{ request.deleteOldDn }}</p>
        </div>
        <div class="col">
            <p class="label">New Superior DN</p>
            <p>{{ request.newSuperiorDn }}</p>
        </div>
    </div>
</template>
