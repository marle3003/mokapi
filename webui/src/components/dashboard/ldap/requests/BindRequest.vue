<script setup lang="ts">
defineProps<{
    request: LdapBindRequest
}>()
</script>

<template>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Name</p>
            <p>{{ request.name }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col-2">
            <p class="label">Auth</p>
            <p>{{ request.auth }}</p>
        </div>
        <div class="col-2">
            <p class="label">Password</p>
            <p>{{ request.password }}</p>
        </div>
    </div>
</template>
