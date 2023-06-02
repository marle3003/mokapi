<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    service: { type: Object as PropType<LdapService>, required: true },
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="row">
                <div class="col">
                    <p class="label">Address</p>
                    <p data-testid="address">{{ service.server }}</p>
                </div>
            </div>
        </div>
    </div>
</template>