<script setup lang="ts">
// The message and the matched DN of a failed response. Both are left out when the server sent none,
// a search response has neither.
defineProps<{
    response?: Partial<LdapResponse>
}>()
</script>

<template>
    <div class="row mb-2" v-if="response?.message">
        <div class="col">
            <p class="label">Message</p>
            <p>{{ response.message }}</p>
        </div>
    </div>
    <div class="row mb-2" v-if="response?.matchedDn">
        <div class="col">
            <p class="label">Matched DN</p>
            <p>{{ response.matchedDn }}</p>
        </div>
    </div>
</template>
