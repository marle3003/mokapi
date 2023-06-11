<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    service: { type: Object as PropType<SmtpService>, required: true },
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
                <div class="col">
                    <p class="label">Max Recipient</p>
                    <p data-testid="max-maxrecipient">{{ !service.maxRecipients ||service.maxRecipients == 0 ? 'unlimited' : service.maxRecipients }}</p>
                </div>
            </div>
        </div>
    </div>
</template>